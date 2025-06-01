package app

import (
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	router *Router
}

func NewHandler() *Handler {
	h := &Handler{
		router: NewRouter(),
	}

	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// GetHandler handles GET reqs
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

// PostHandler handles POsT reqs
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
