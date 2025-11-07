package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/JohnnyConstantin/urlshort/internal/store"
)

// TestRouterServeHTTP проверяет обработку запросов роутером
func TestRouterServeHTTP(t *testing.T) {
	var s Server
	service := Service{}
	server := s.NewServer(&service)
	router := server.Router

	testHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			return
		}
	}

	router.AddRoute("/test", http.MethodGet, testHandler)

	req, err := http.NewRequest(http.MethodGet, "/test", nil)
	assert.NoErrorf(t, err, "Expected no error for GET request")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	req, err = http.NewRequest(http.MethodGet, "/nonexistent", nil)
	assert.NoErrorf(t, err, "Expected no error for GET request")

	rr = httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equalf(t,
		store.DefaultErrorCode,
		rr.Code,
		"handler returned wrong status code: got %v want %v", rr.Code, store.DefaultErrorCode)
}

// TestRouterAddRoute проверяет добавление маршрутов в роутер
func TestRouterAddRoute(t *testing.T) {
	var s Server
	service := Service{}
	server := s.NewServer(&service)
	router := server.Router

	// Добавляем тестовый маршрут
	testHandler := func(w http.ResponseWriter, r *http.Request) {}
	router.AddRoute("/test", http.MethodGet, testHandler)

	// Проверяем, что маршрут добавлен
	assert.NotEmptyf(t, router, "Router should not be empty")

	//Проверяем, что создался указанный маршрут
	route := router.Routes["/test"]
	assert.NotNil(t, route, "Expected route to exist")

	assert.NotEmptyf(t, route[http.MethodGet], "Expected handler for method GET")
}
