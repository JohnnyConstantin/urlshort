package main

import (
	"database/sql"
	route "github.com/go-chi/chi/v5"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/JohnnyConstantin/urlshort/internal/app"
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

func TestCreateHandlers(t *testing.T) {
	// Создаем тестовый логгер
	logger := zaptest.NewLogger(t)
	sugar := *logger.Sugar()

	// Создаем mock базу данных
	db := &sql.DB{} // Пустая структура, так как мы только проверяем маршруты

	// Создаем тестовый handler
	var s app.Server
	server := s.NewServer()
	handler := server.Handler

	router := route.NewRouter()

	// Вызываем тестируемую функцию
	createHandlers(db, router, sugar, handler)

	// Тестируем наличие ожидаемых маршрутов
	testCases := []struct {
		method string
		path   string
	}{
		{"POST", "/"},
		{"POST", "/api/shorten"},
		{"POST", "/api/shorten/batch"},
		{"DELETE", "/api/user/urls"},
		{"GET", "/api/user/urls"},
		{"GET", "/{id}"},
		{"GET", "/ping"},
	}

	for _, tc := range testCases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rr := httptest.NewRecorder()

			router.ServeHTTP(rr, req)

			// Мы не проверяем статус код, так как он зависит от реализации хендлеров
			// Главное - что маршрут существует и не возвращает 404
			if rr.Code == http.StatusNotFound {
				t.Errorf("Route %s %s not found", tc.method, tc.path)
			}
		})
	}
}
