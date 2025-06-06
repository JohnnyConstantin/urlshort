package app

import (
	"bytes"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestPostHandler проверяет хендлер POST запросов
func TestPostHandler(t *testing.T) {
	var s Server
	server := s.NewServer()
	handler := server.Handler

	requestBody := "https://example.com"
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.PostHandler(rr, req)

	// Проверка статуса (201)
	assert.Equalf(
		t,
		rr.Code,
		http.StatusCreated,
		"handler returned wrong status code: got %v want %v", rr.Code, http.StatusCreated)

	// Проверка возвращения shortenurl (только наличие, не сам формат)
	response := rr.Body.Bytes()
	assert.NotEmptyf(t, response, "Response should not be empty")
}

// TestGetHandler проверяет хендлер GET запросов
func TestGetHandler(t *testing.T) {
	testURL := "https://example.com"
	var s Server
	server := s.NewServer()
	handler := server.Handler

	// Создание короткой ссылки (на случай, если тест запускается атомарно, без TestPostHandler)
	requestBody := testURL
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(requestBody))
	require.NoErrorf(t, err, "Expected no error for POST request")

	rr := httptest.NewRecorder()
	handler.PostHandler(rr, req)

	println(rr.Body.String())

	shortURL := rr.Body.Bytes()
	id := strings.TrimPrefix(string(shortURL), config.Options.BaseAddress)

	// Проверка Get обработчика
	req, err = http.NewRequest(http.MethodGet, "/"+id, nil)
	require.NoErrorf(t, err, "Expected no error for GET request")

	rr = httptest.NewRecorder()
	handler.GetHandler(rr, req)

	assert.Equalf(t,
		http.StatusTemporaryRedirect,
		rr.Code,
		"GET handler returned wrong status code: %v, expected %v", rr.Code, http.StatusTemporaryRedirect)

	location := rr.Header().Get("Location")
	assert.Equalf(t, testURL, location, "expected Location header %s, got '%s'", testURL, location)
}
