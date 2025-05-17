package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	log "github.com/sirupsen/logrus"
)

// EnrichResult представляет результат обогащения данных.
type EnrichResult struct {
	Age         *int    `json:"age"`
	Gender      *string `json:"gender"`
	Nationality *string `json:"nationality"`
}

// Enrich обогащает данные о человеке, используя API Agify, Genderize и Nationalize.
// Возвращает результат обогащения и ошибку, если таковая возникла.
func Enrich(name string) (*EnrichResult, error) {
	log.Debugf("service.Enrich: starting enrichment for name=%s", name)

	var (
		wg  sync.WaitGroup
		res EnrichResult
		mu  sync.Mutex
		err error
	)

	wg.Add(3)

	// Agify
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.agify.io/?name=%s", name)
		log.Debugf("service.Enrich: calling Agify API: %s", url)

		var a struct {
			Age *int `json:"age"`
		}
		if e := callAPI(url, &a); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			log.WithError(e).Error("service.Enrich: Agify call failed")
			return
		}

		mu.Lock()
		res.Age = a.Age
		mu.Unlock()
		log.Debugf("service.Enrich: Agify result: %v", a.Age)
	}()

	// Genderize
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.genderize.io/?name=%s", name)
		log.Debugf("service.Enrich: calling Genderize API: %s", url)

		var g struct {
			Gender *string `json:"gender"`
		}
		if e := callAPI(url, &g); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			log.WithError(e).Error("service.Enrich: Genderize call failed")
			return
		}

		mu.Lock()
		res.Gender = g.Gender
		mu.Unlock()
		log.Debugf("service.Enrich: Genderize result: %v", g.Gender)
	}()

	// Nationalize
	go func() {
		defer wg.Done()
		url := fmt.Sprintf("https://api.nationalize.io/?name=%s", name)
		log.Debugf("service.Enrich: calling Nationalize API: %s", url)

		var n struct {
			Country []struct {
				CountryID   string  `json:"country_id"`
				Probability float64 `json:"probability"`
			} `json:"country"`
		}
		if e := callAPI(url, &n); e != nil {
			mu.Lock()
			err = e
			mu.Unlock()
			log.WithError(e).Error("service.Enrich: Nationalize call failed")
			return
		}

		if len(n.Country) > 0 {
			mu.Lock()
			res.Nationality = &n.Country[0].CountryID
			mu.Unlock()
			log.Debugf("service.Enrich: Nationalize result: %v", n.Country[0].CountryID)
		}
	}()

	wg.Wait()

	if err != nil {
		log.WithError(err).Error("service.Enrich: enrichment failed")
		return nil, err
	}

	log.Debugf("service.Enrich: completed enrichment for name=%s: %+v", name, res)
	return &res, nil
}

// callAPI отправляет GET запрос на указанный URL и декодирует ответ в указанный интерфейс.
// Если тело ответа пустое (EOF), ошибка игнорируется.
func callAPI(url string, out interface{}) error {
	log.Debugf("service.callAPI: GET %s", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		// если тело было пустое — игнорируем
		if err == io.EOF {
			return nil
		}
		return err
	}
	return nil
}
