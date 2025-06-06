package main

import (
	_ "encoding/json"
	"flag"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
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
	router.Post("/api/shorten", handler.PostHandler)
	router.Get("/{id}", handler.GetHandler)

	flag.Parse()

	err := http.ListenAndServe(config.Options.Address, router)
	if err != nil {
		return
	}

}
