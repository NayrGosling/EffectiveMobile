package service

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

// rewriteTransport перенаправляет запросы к разным хостам на соответствующие тестовые серверы.
type rewriteTransport struct {
	base    http.RoundTripper
	mapping map[string]string
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Оригинальный хост, по которому строился URL
	origHost := req.URL.Host
	if newHost, ok := t.mapping[origHost]; ok {
		req.URL.Scheme = "http"
		req.URL.Host = newHost
	}
	return t.base.RoundTrip(req)
}

// TestCallAPI_Success тестирует функцию callAPI с корректным JSON ответом.
// Ожидается, что функция успешно декодирует JSON и вернет ожидаемый результат.
func TestCallAPI_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(struct {
			Foo string `json:"foo"`
		}{Foo: "bar"})
	}))
	defer srv.Close()

	var out struct{ Foo string }
	if err := callAPI(srv.URL, &out); err != nil {
		t.Fatalf("callAPI returned error: %v", err)
	}

	if out.Foo != "bar" {
		t.Errorf("expected Foo=bar, got %q", out.Foo)
	}
}

// TestCallAPI_NonJSON тестирует функцию callAPI с некорректным JSON ответом.
// Ожидается, что функция вернет ошибку при попытке декодирования некорректного JSON.
func TestCallAPI_NonJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	var out struct{ Foo string }
	err := callAPI(srv.URL, &out)

	if err == nil {
		t.Fatal("expected error decoding non-JSON, got nil")
	}
}

// TestEnrich_Success тестирует функцию Enrich в случае успешного выполнения.
// Он создает тестовые серверы для agify, genderize и nationalize, которые имитируют ответы API.
// Затем он настраивает http.DefaultClient для использования этих тестовых серверов.
// Наконец, он вызывает функцию Enrich и проверяет, что она возвращает ожидаемые результаты.
func TestEnrich_Success(t *testing.T) {
	agify := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"age":30}`)
	}))
	defer agify.Close()

	genderize := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"gender":"male"}`)
	}))
	defer genderize.Close()

	nationalize := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"country":[{"country_id":"US","probability":0.9}]}`)
	}))
	defer nationalize.Close()

	origClient := http.DefaultClient
	http.DefaultClient = &http.Client{
		Transport: &rewriteTransport{
			base: http.DefaultTransport,
			mapping: map[string]string{
				"api.agify.io":       mustHostPort(agify.URL),
				"api.genderize.io":   mustHostPort(genderize.URL),
				"api.nationalize.io": mustHostPort(nationalize.URL),
			},
		},
		Timeout: 2 * time.Second,
	}
	defer func() { http.DefaultClient = origClient }()

	res, err := Enrich("john")
	if err != nil {
		t.Fatalf("Enrich returned error: %v", err)
	}
	if res.Age == nil || *res.Age != 30 {
		t.Errorf("expected Age=30, got %v", res.Age)
	}
	if res.Gender == nil || *res.Gender != "male" {
		t.Errorf("expected Gender=male, got %v", res.Gender)
	}
	if res.Nationality == nil || *res.Nationality != "US" {
		t.Errorf("expected Nationality=US, got %v", res.Nationality)
	}
}

// TestEnrich_PartialFailure тестирует функцию Enrich в случае частичного сбоя.
// Он создает тестовые серверы для agify, genderize и nationalize, которые имитируют ответы API.
// Затем он настраивает http.DefaultClient для использования этих тестовых серверов.
// Наконец, он вызывает функцию Enrich и проверяет, что она возвращает ошибку, когда один из API возвращает 500.
func TestEnrich_PartialFailure(t *testing.T) {
	agify := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	defer agify.Close()

	genderize := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"gender":"female"}`)
	}))
	defer genderize.Close()

	nationalize := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"country":[{"country_id":"CA","probability":0.5}]}`)
	}))
	defer nationalize.Close()

	http.DefaultClient = &http.Client{
		Transport: &rewriteTransport{
			base: http.DefaultTransport,
			mapping: map[string]string{
				"api.agify.io":       mustHostPort(agify.URL),
				"api.genderize.io":   mustHostPort(genderize.URL),
				"api.nationalize.io": mustHostPort(nationalize.URL),
			},
		},
		Timeout: 2 * time.Second,
	}

	_, err := Enrich("alice")
	if err == nil {
		t.Fatal("expected error when one of APIs вернул 500, got nil")
	}
}

// mustHostPort возвращает хост и порт из URL в виде строки.
// Если URL некорректен или не содержит порт, функция вызывает панику.
func mustHostPort(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}

	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		panic(err)
	}

	return net.JoinHostPort(host, port)
}
