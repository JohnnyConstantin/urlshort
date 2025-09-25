package app

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
)

// TestPostHandler проверяет хендлер POST запросов
func TestPostHandler(t *testing.T) {
	var s Server
	server := s.NewServer()
	handler := server.Handler

	config.CreateStorageConfig()
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

func BenchmarkPostHandler(t *testing.B) {
	t.StopTimer() // останавливаем таймер
	var s Server
	server := s.NewServer()
	handler := server.Handler

	config.CreateStorageConfig()
	requestBody := "https://example.com"
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	t.StartTimer() // Запускаем таймер после подготовки данных
	handler.PostHandler(rr, req)
}

// TestGetHandler проверяет хендлер GET запросов
func TestGetHandler(t *testing.T) {
	config.CreateStorageConfig()
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

	shortURL := rr.Body.Bytes()
	id := strings.TrimPrefix(string(shortURL), config.Options.BaseAddress+"/")

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

func BenchmarkGetHandler(t *testing.B) {
	t.StopTimer() // останавливаем таймер
	config.CreateStorageConfig()
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

	shortURL := rr.Body.Bytes()
	id := strings.TrimPrefix(string(shortURL), config.Options.BaseAddress+"/")

	// Проверка Get обработчика
	req, err = http.NewRequest(http.MethodGet, "/"+id, nil)
	require.NoErrorf(t, err, "Expected no error for GET request")

	rr = httptest.NewRecorder()
	t.StartTimer() // Возобновляем таймер перед запуском GET хендлера
	handler.GetHandler(rr, req)
}

func TestPostHandlerMultiple(t *testing.T) {
	var s Server
	server := s.NewServer()
	handler := server.Handler

	tests := []struct {
		name           string
		requestBody    string
		storageType    config.StorageType
		expectedStatus int
		wantError      bool
		setupContext   func(r *http.Request) *http.Request
	}{
		{
			name: "Successful batch request with memory storage",
			requestBody: `[
				{"correlation_id": "1", "original_url": "https://example.com/1"},
				{"correlation_id": "2", "original_url": "https://example.com/2"}
			]`,
			storageType:    config.StorageMemory,
			expectedStatus: http.StatusCreated,
			wantError:      false,
		},
		{
			name: "Successful batch request with file storage",
			requestBody: `[
				{"correlation_id": "1", "original_url": "https://example.com/1"},
				{"correlation_id": "2", "original_url": "https://example.com/2"}
			]`,
			storageType:    config.StorageFile,
			expectedStatus: http.StatusCreated,
			wantError:      false,
		},
		{
			name:           "Invalid JSON format",
			requestBody:    `invalid json`,
			storageType:    config.StorageMemory,
			expectedStatus: store.DefaultErrorCode,
			wantError:      true,
		},
		{
			name:           "Request body too large",
			requestBody:    strings.Repeat("a", 1024*1024+1), // > 1MB
			storageType:    config.StorageMemory,
			expectedStatus: store.DefaultErrorCode,
			wantError:      true,
		},
		{
			name: "Missing original_url in request",
			requestBody: `[
				{"correlation_id": "1", "original_url": "https://example.com/1"},
				{"correlation_id": "2"}
			]`,
			storageType:    config.StorageMemory,
			expectedStatus: http.StatusCreated,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем тип хранилища
			config.CreateStorageConfig()

			req, err := http.NewRequest(http.MethodPost, "/api/shorten/batch", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.PostHandlerMultiple(rr, req)

			// Проверка статуса
			assert.Equalf(t, tt.expectedStatus, rr.Code,
				"handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)

			if !tt.wantError {
				// Проверка успешного ответа
				assert.Equalf(t, http.StatusCreated, rr.Code,
					"handler returned wrong status code for success case: got %v want %v", rr.Code, http.StatusCreated)

				// Проверка Content-Type
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

				// Проверка структуры ответа
				var responses []models.BatchShortenResponse
				err := json.Unmarshal(rr.Body.Bytes(), &responses)
				assert.NoError(t, err, "Response should be valid JSON")

				// Для непустых запросов проверяем соответствие количества элементов
				if tt.requestBody != "[]" {
					var requests []models.BatchShortenRequest
					json.Unmarshal([]byte(tt.requestBody), &requests)
					assert.Equal(t, len(requests), len(responses), "Number of responses should match requests")

					// Проверяем, что correlation_id сохранились
					for i, resp := range responses {
						assert.Equal(t, requests[i].CorrelationID, resp.CorrelationID)
						assert.NotEmpty(t, resp.ShortURL, "ShortURL should not be empty")
					}
				} else {
					// Для пустого батча должен вернуться пустой массив
					assert.Equal(t, 0, len(responses), "Response should be empty array")
				}
			} else {
				// Проверка ошибки - тело должно содержать сообщение об ошибке
				assert.NotEmpty(t, rr.Body.String(), "Error response should not be empty")
			}
		})
	}
}

func BenchmarkPostHandlerMultiple(t *testing.B) {
	t.StopTimer() // останавливаем таймер
	var s Server
	server := s.NewServer()
	handler := server.Handler

	config.CreateStorageConfig()
	requestBody := `[
    {
        "correlation_id": "asdxadcv",
        "original_url": "www.sodsbiubidfsfcmefvfvurlsd.com"
    },
    {
        "correlation_id": "asdsdfsffff",
        "original_url": "www.sosdfjniundmeurldccfsdfsdffvfvsd.com"
    }
]`
	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	t.StartTimer() // Запускаем таймер после подготовки данных
	handler.PostHandlerMultiple(rr, req)
}
