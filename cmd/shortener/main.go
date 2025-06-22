package main

import (
	"encoding/json"
	"flag"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/models"
	route "github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
	"os"
)

func main() {
	var s app.Server
	server := s.NewServer()

	//Чтобы удобнее было работать
	handler := server.Handler
	router := route.NewRouter() //Используем внешний роутер chi, вместо встроенного в объект app.Server

	//Накидываем хендлеры на роуты
	router.Route("/", func(r route.Router) {
		r.Post("/", handler.PostHandler)
		r.Route("/api", func(r route.Router) {
			r.Post("/shorten", jsonResponseMiddleware(handler.PostHandler))
		})
		r.Get("/{id}", handler.GetHandler)
	})

	flag.Parse()

	if envA := os.Getenv("SERVER_ADDRESS"); envA != "" {
		config.Options.Address = envA
	}
	if envB := os.Getenv("BASE_URL"); envB != "" {
		config.Options.BaseAddress = envB
	}

	err := http.ListenAndServe(config.Options.Address, router)
	if err != nil {
		return
	}

}

// Middleware для проверки того, что возвращаемое значение является JSON. Иначе переводит его в JSON
func jsonResponseMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rr := httptest.NewRecorder()
		next(rr, r)

		for k, v := range rr.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rr.Code)

		if rr.Header().Get("Content-Type") == "application/json" {
			var response models.ShortenResponse
			response.Result = rr.Body.String()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}
	}
}
