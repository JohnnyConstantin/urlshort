package main

import (
	"bytes"
	"encoding/json"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestServerCreation проверяет корректность создания сервера
func TestServerCreation(t *testing.T) {
	var s app.Server
	server := s.NewServer()
	router := server.Router
	handler := server.Handler

	require.NotNilf(t, server, "Expected server instance, got nil")

	require.NotNilf(t, router, "Expected router to be initialized, got nil")

	require.NotNilf(t, handler, "Expected handler to be initialized, got nil")
}

func TestPostHandler(t *testing.T) {
	var s app.Server
	server := s.NewServer()
	handler := server.Handler

	requestBody := "https://example.com"
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.PostHandler(rr, req)

	// Проверка что возвращается json
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
	}

	// Проверка что возвращаемое значение является валидным JSON
	var jsonResponse map[string]interface{}
	if err := json.Unmarshal(rr.Body.Bytes(), &jsonResponse); err != nil {
		t.Errorf("Response is not valid JSON: %v", err)
	}
}
