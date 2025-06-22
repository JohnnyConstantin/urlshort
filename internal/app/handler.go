package app

import (
	"compress/gzip"
	"encoding/json"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	_ "github.com/lib/pq"
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
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 1 {
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	id := parts[0]

	response, exists := getFullURL(id)
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

	ShortURL = shortenURL(OriginalURL.URL)

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ShortURL.Result))
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
