package app

import (
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
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}
	defer r.Body.Close()

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
