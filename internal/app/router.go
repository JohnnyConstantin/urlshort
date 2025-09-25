package app

import (
	"net/http"
	"strings"

	"github.com/JohnnyConstantin/urlshort/internal/store"
)

type Router struct {
	Routes map[string]map[string]http.HandlerFunc
}

// NewRouter возвращает новый *Router с пустыми роутами
func NewRouter() *Router {
	return &Router{
		Routes: make(map[string]map[string]http.HandlerFunc), //Словарь, содержащий роуты,методы и хендлеры к ним
	}
}

// AddRoute регистрирует новый хендлер для роута. Если передается существующий роут - он перезаписывается
func (r *Router) AddRoute(path string, method string, handler http.HandlerFunc) {
	if _, ok := r.Routes[path]; !ok {
		r.Routes[path] = make(map[string]http.HandlerFunc)
	}
	r.Routes[path][method] = handler
}

// ServeHTTP Роутит запросы на зарезервированный в структуре Router хендлер. Необходим для роутера,
// чтобы подстроиться под стандартную либу net.http с помощью утиной типизации
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	method := req.Method

	if methods, ok := r.Routes[path]; ok {
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
			if methods, oks := r.Routes["/{id}"]; oks {
				if handler, ok := methods[http.MethodGet]; ok {
					handler(w, req)
					return
				}
			}
		}
	}

	http.Error(w, store.DefaultError, store.DefaultErrorCode)
}
