package app

import (
	"compress/gzip"
	"encoding/json"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	_ "github.com/jackc/pgx/v5/stdlib"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	router *Router
}

// NewHandler Инциализация объекта хендлера с пустым роутером
func NewHandler() *Handler {
	h := &Handler{
		router: NewRouter(),
	}

	return h
}

// ServeHTTP Утиная типизация, прокидываемся до функциональной части роутера по роутингу запросов на хендлер
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// GetHandler обрабатывает GET запросы
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	response := models.ShortenRequest{URL: ""}
	exists := false //By default не существует
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 1 {
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	id := parts[0]

	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		fuller := FileFuller{cfg}
		response, exists = fuller.GetFullURL(id)
	case config.StorageMemory:
		fuller := MemoryFuller{cfg}
		response, exists = fuller.GetFullURL(id)
	case config.StorageDB:
		fuller := DBFuller{cfg}
		response, exists = fuller.GetFullURL(id)
	default: // Overkill, но перестраховаться нужно
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
	}

	if !exists {
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
	}

	result := response.URL

	w.Header().Set("Location", result)

	w.WriteHeader(307)
}

// PostHandler обрабатывает POST запросы
func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	var ShortURL models.ShortenResponse
	var OriginalURL models.ShortenRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, store.ReadBodyError, store.DefaultErrorCode)
		return
	}
	defer r.Body.Close()

	// Доп. обработка тела
	// Ограничиваем размер тела запроса
	if len(body) > 1024*1024 { // 1MB
		http.Error(w, store.LargeBodyError, store.DefaultErrorCode)
		return
	}

	if r.Header.Get("Content-Type") == "application/json" {
		if err = json.Unmarshal(body, &OriginalURL); err != nil {
			http.Error(w, store.DefaultError, store.DefaultErrorCode)
		}
		w.Header().Set("Content-Type", "application/json")
	} else {
		OriginalURL.URL = string(body)
		w.Header().Set("Content-Type", "text/plain")
	}

	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		shortener := FileShortener{cfg}
		ShortURL = shortener.ShortenURL(OriginalURL.URL)
	case config.StorageMemory:
		shortener := MemoryShortener{cfg}
		ShortURL = shortener.ShortenURL(OriginalURL.URL)
	case config.StorageDB:
		shortener := DBShortener{cfg}
		ShortURL = shortener.ShortenURL(OriginalURL.URL)
	default: // Overkill, но перестраховаться нужно
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ShortURL)
}

// PostHandlerMultiple обрабатывает POST запросы с batch
func (h *Handler) PostHandlerMultiple(w http.ResponseWriter, r *http.Request) {
	var requests []models.BatchShortenRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, store.ReadBodyError, store.DefaultErrorCode)
		return
	}
	defer r.Body.Close()

	// Доп. обработка тела
	// Ограничиваем размер тела запроса
	if len(body) > 1024*1024 { // 1MB
		http.Error(w, store.LargeBodyError, store.DefaultErrorCode)
		return
	}

	if err := json.Unmarshal(body, &requests); err != nil {
		http.Error(w, "Invalid batch request format", store.DefaultErrorCode)
		return
	}

	responses := make([]models.BatchShortenResponse, 0, len(requests))
	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		shortener := FileShortener{cfg}
		for _, req := range requests {
			ShortURL := shortener.ShortenURL(req.OriginalURL)

			responses = append(responses, models.BatchShortenResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      ShortURL.Result,
			})
		}

	case config.StorageMemory:
		shortener := MemoryShortener{cfg}
		for _, req := range requests {
			ShortURL := shortener.ShortenURL(req.OriginalURL)

			responses = append(responses, models.BatchShortenResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      ShortURL.Result,
			})
		}
	case config.StorageDB:
		shortener := DBShortener{cfg}
		for _, req := range requests {
			ShortURL := shortener.ShortenURL(req.OriginalURL)

			responses = append(responses, models.BatchShortenResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      ShortURL.Result,
			})
		}
	default: // Overkill, но перестраховаться нужно
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func GzipHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверка на то, что клиент прислал пожатый контент
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body) // Распаковываем
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		originalWriter := w
		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") // Проверяем что клиент поддерживает сжатие

		if acceptsGzip { // Если клиент поддерживает сжатие, проверяем передаваемый content-type
			contentType := r.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "application/json") ||
				strings.HasPrefix(contentType, "text/html") {

				gzWriter := gzip.NewWriter(w) // Жмем!
				defer gzWriter.Close()

				w.Header().Set("Content-Encoding", "gzip") // Ставим заголовок, что пожали контент
				originalWriter = &gzipWriter{
					ResponseWriter: w,
					Writer:         gzWriter,
				}
			}
		}

		next(originalWriter, r) // перекидываем дальше
	}
}

// PingDBHandler Проверяет подключение к БД
func (h *Handler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	dsn := config.Options.DSN
	dbConn, err := GetDBConnection(dsn)
	if err != nil {
		http.Error(w, store.ConnectionError, store.InternalSeverErrorCode)
	}

	defer dbConn.Close()

	err = dbConn.Ping()
	if err != nil {
		http.Error(w, store.ConnectionError, store.InternalSeverErrorCode)
	}

	w.WriteHeader(http.StatusOK)
}
