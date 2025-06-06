package main

import (
	"encoding/json"
	"flag"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	route "github.com/go-chi/chi/v5"
	"net/http"
	"net/http/httptest"
)

func main() {
	var s app.Server
	server := s.NewServer()

	//Чтобы удобнее было работать
	handler := server.Handler
	router := route.NewRouter() //Используем внешний роутер chi, вместо встроенного в объект app.Server

	//Накидываем хендлеры на роуты
	router.Post("/", jsonResponseMiddleware(handler.PostHandler))
	router.Get("/{id}", handler.GetHandler)

	flag.Parse()

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
			w.Write(rr.Body.Bytes())
			return
		}

		response := map[string]string{
			"url": rr.Body.String(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}
