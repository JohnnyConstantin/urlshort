package app

import (
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
	var response models.URLResponse
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

	w.Header().Set("Location", response.OriginalURL)

	w.WriteHeader(307)
}

// PostHandler обрабатывает POST запросы
func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	var LongURL models.ShortenRequest
	var ShortURL models.ShortenResponse

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	LongURL.URL = string(body)
	ShortURL = shortenURL(LongURL.URL)

	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://" + r.Host + "/" + ShortURL.Result))
}
