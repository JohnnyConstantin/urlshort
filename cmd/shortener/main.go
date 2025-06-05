package main

import (
	"github.com/JohnnyConstantin/urlshort/internal/app"
	route "github.com/go-chi/chi/v5"
	"net/http"
)

func main() {
	var s app.Server
	server := s.NewServer()

	//Чтобы удобнее было работать
	handler := server.Handler
	router := route.NewRouter() //Используем внешний роутер chi, вместо встроенного в объект app.Server

	//Накидываем хендлеры на роуты
	router.Post("/", handler.PostHandler)
	router.Get("/{id}", handler.GetHandler)

	err := http.ListenAndServe(":8080", router)
	if err != nil {
		return
	}

}
