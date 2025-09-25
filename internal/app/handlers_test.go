package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
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

// Вспомогательная функция для создания gzip сжатых данных
func gzipData(data string) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Тестовый хендлер для проверки
func testHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(body) // Эхо-ответ с телом запроса
}

func TestGzipHandle(t *testing.T) {
	tests := []struct {
		name                string
		contentEncoding     string
		acceptEncoding      string
		contentType         string
		requestBody         string
		compressRequestBody bool
		expectGzipResponse  bool
		expectError         bool
		expectedStatusCode  int
	}{
		{
			name:                "Gzip request and response",
			contentEncoding:     "gzip",
			acceptEncoding:      "gzip",
			contentType:         "application/json",
			requestBody:         `{"test": "data"}`,
			compressRequestBody: true,
			expectGzipResponse:  true,
			expectedStatusCode:  http.StatusOK,
		},
		{
			name:                "Gzip request without gzip acceptance",
			contentEncoding:     "gzip",
			acceptEncoding:      "",
			contentType:         "application/json",
			requestBody:         `{"test": "data"}`,
			compressRequestBody: true,
			expectGzipResponse:  false,
			expectedStatusCode:  http.StatusOK,
		},
		{
			name:                "Non-gzip request with gzip acceptance",
			contentEncoding:     "",
			acceptEncoding:      "gzip",
			contentType:         "application/json",
			requestBody:         `{"test": "data"}`,
			compressRequestBody: false,
			expectGzipResponse:  true,
			expectedStatusCode:  http.StatusOK,
		},
		{
			name:                "Invalid gzip body",
			contentEncoding:     "gzip",
			acceptEncoding:      "gzip",
			contentType:         "application/json",
			requestBody:         "invalid gzip data",
			compressRequestBody: false, // Отправляем не сжатые данные с заголовком gzip
			expectError:         true,
			expectedStatusCode:  http.StatusBadRequest,
		},
		{
			name:                "Gzip acceptance but non-compressible content type",
			contentEncoding:     "",
			acceptEncoding:      "gzip",
			contentType:         "image/png",
			requestBody:         "binary data",
			compressRequestBody: false,
			expectGzipResponse:  false,
			expectedStatusCode:  http.StatusOK,
		},
		{
			name:                "Text HTML with gzip",
			contentEncoding:     "",
			acceptEncoding:      "gzip",
			contentType:         "text/html",
			requestBody:         "<html><body>Test</body></html>",
			compressRequestBody: false,
			expectGzipResponse:  true,
			expectedStatusCode:  http.StatusOK,
		},
		{
			name:                "No compression needed",
			contentEncoding:     "",
			acceptEncoding:      "",
			contentType:         "application/json",
			requestBody:         `{"test": "data"}`,
			compressRequestBody: false,
			expectGzipResponse:  false,
			expectedStatusCode:  http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготовка тела запроса
			var body io.Reader
			if tt.compressRequestBody {
				compressed, err := gzipData(tt.requestBody)
				if err != nil {
					t.Fatalf("Failed to compress test data: %v", err)
				}
				body = bytes.NewReader(compressed)
			} else {
				body = strings.NewReader(tt.requestBody)
			}

			// Создание запроса
			req := httptest.NewRequest("POST", "/test", body)
			req.Header.Set("Content-Encoding", tt.contentEncoding)
			req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			req.Header.Set("Content-Type", tt.contentType)

			// Создание ResponseRecorder
			rr := httptest.NewRecorder()

			// Создание middleware с тестовым хендлером
			handler := GzipHandle(testHandler)

			// Выполнение запроса
			handler.ServeHTTP(rr, req)

			// Проверка статус кода
			if rr.Code != tt.expectedStatusCode {
				t.Errorf("Expected status code %d, got %d", tt.expectedStatusCode, rr.Code)
			}

			// Если ожидалась ошибка, проверяем сообщение
			if tt.expectError {
				if !strings.Contains(rr.Body.String(), "Invalid gzip body") {
					t.Errorf("Expected error message 'Invalid gzip body', got: %s", rr.Body.String())
				}
				return
			}

			// Проверка заголовков ответа
			contentEncoding := rr.Header().Get("Content-Encoding")
			if tt.expectGzipResponse {
				if contentEncoding != "gzip" {
					t.Errorf("Expected Content-Encoding: gzip, got: %s", contentEncoding)
				}

				// Проверяем, что ответ действительно сжат
				if strings.Contains(contentEncoding, "gzip") {
					// Пытаемся распаковать ответ
					gr, err := gzip.NewReader(rr.Body)
					if err != nil {
						t.Errorf("Response is not valid gzip: %v", err)
					}
					defer gr.Close()

					uncompressed, err := io.ReadAll(gr)
					if err != nil {
						t.Errorf("Failed to decompress response: %v", err)
					}

					// Проверяем, что распакованные данные совпадают с исходными
					if string(uncompressed) != tt.requestBody {
						t.Errorf("Decompressed response doesn't match expected. Expected: %s, Got: %s",
							tt.requestBody, string(uncompressed))
					}
				}
			} else {
				if contentEncoding != "" {
					t.Errorf("Expected no Content-Encoding, got: %s", contentEncoding)
				}

				// Проверяем, что ответ не сжат
				if rr.Body.String() != tt.requestBody {
					t.Errorf("Response body doesn't match expected. Expected: %s, Got: %s",
						tt.requestBody, rr.Body.String())
				}
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
