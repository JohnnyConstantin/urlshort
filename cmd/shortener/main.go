package main

import (
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"net/http"
)

func main() {
	var s app.Server
	server := s.NewServer()

	//Чтобы удобнее было работать
	handler := server.Handler
	router := server.Router

	//Накидываем хендлеры на роуты
	router.AddRoute("/", http.MethodPost, handler.PostHandler)
	router.AddRoute("/{id}", http.MethodGet, handler.GetHandler)

	//Регаем handlers через стандартную либу
	http.Handle("/", router)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}

}
