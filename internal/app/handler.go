package app

import (
	"net/http"
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
	w.Write([]byte("Это GET хендлер"))
}

// PostHandler handles POsT reqs
func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Это POST хендлер"))
}
