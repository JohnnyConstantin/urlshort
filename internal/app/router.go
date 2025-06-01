package app

import (
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"net/http"
	"strings"
)

type Router struct {
	routes map[string]map[string]http.HandlerFunc
}

// NewRouter returns new *Router with empty routes
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

// AddRoute registers new route for handler. If passed existing one, it is overwritten
func (r *Router) AddRoute(path string, method string, handler http.HandlerFunc) {
	if _, ok := r.routes[path]; !ok {
		r.routes[path] = make(map[string]http.HandlerFunc)
	}
	r.routes[path][method] = handler
}

// ServeHTTP Routes request to handler method
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	if methods, ok := r.routes[path]; ok {
		if handler, oks := methods[method]; oks {
			handler(w, req)
			return
		}
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	if method == http.MethodGet && path != "/" {
		trimmedPath := strings.Trim(path, "/")
		if trimmedPath != "" && !strings.Contains(trimmedPath, "/") {
			if methods, oks := r.routes["/{id}"]; oks {
				if handler, ok := methods[http.MethodGet]; ok {
					handler(w, req)
					return
				}
			}
		}
	}

	http.Error(w, store.DefaultError, store.DefaultErrorCode)
}
