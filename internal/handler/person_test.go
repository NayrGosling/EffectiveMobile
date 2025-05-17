package handler

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Заглушка для sql.DB
type fakeDB struct{}

func (f *fakeDB) QueryRow(query string, args ...interface{}) *sql.Row {
	// Возвращаем sql.ErrNoRows
	return &sql.Row{}
}
func (f *fakeDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("exec error")
}

// TestGetByID_NotFound проверяет, что обработчик GetByID возвращает статус 404, когда запись не найдена.
func TestGetByID_NotFound(t *testing.T) {
	h := &PersonHandler{DB: (*sql.DB)(nil)}
	req := httptest.NewRequest(http.MethodGet, "/persons/123", nil)
	rw := httptest.NewRecorder()
	h.GetByID(rw, req)
	if rw.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rw.Code)
	}
}

// TestCreate_BadPayload проверяет, что обработчик Create возвращает статус 400, когда получает некорректный JSON.
func TestCreate_BadPayload(t *testing.T) {
	h := &PersonHandler{DB: (*sql.DB)(nil)}
	req := httptest.NewRequest(http.MethodPost, "/persons", bytes.NewBufferString(`{invalid json}`))
	rw := httptest.NewRecorder()
	h.Create(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rw.Code)
	}
}
