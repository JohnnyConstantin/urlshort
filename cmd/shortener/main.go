package main

import (
	"encoding/json"
	"flag"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/models"
	route "github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"os"
)

var sugar zap.SugaredLogger

func main() {
	var s app.Server
	server := s.NewServer()

	//Чтобы удобнее было работать
	handler := server.Handler
	router := route.NewRouter() //Используем внешний роутер chi, вместо встроенного в объект app.Server

	//Создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	sugar = *logger.Sugar() // Создали экземпляр и в дальнейшем прокидываем его в middleware с логированием

	//Накидываем хендлеры на роуты
	router.Route("/", func(r route.Router) {
		r.Post("/",
			app.GzipHandle( // Сжатие
				app.WithLogging( // Логирование, прокидываем в него регистратор логов sugar
					handler.PostHandler, sugar))) // Сам хендлер
		r.Route("/api", func(r route.Router) {
			r.Post("/shorten",
				app.GzipHandle( // Сжатие
					app.WithLogging( // Логирование, прокидываем в него регистратор логов sugar
						jsonResponseMiddleware( // Работа с json request/response
							handler.PostHandler), sugar))) // Сам хендлер
		})
		r.Get("/{id}",
			app.GzipHandle( // Сжатие
				app.WithLogging( // Логирование, прокидываем в него регистратор логов sugar
					handler.GetHandler, sugar))) // Сам хендлер
	})

	flag.Parse()

	//Подгружаем переменные окружения при наличии
	if envA := os.Getenv("SERVER_ADDRESS"); envA != "" {
		config.Options.Address = envA
	}
	if envB := os.Getenv("BASE_URL"); envB != "" {
		config.Options.BaseAddress = envB
	}
	if envC := os.Getenv("FILE_STORAGE_PATH"); envC != "" {
		config.Options.FileToWrite = envC
	}

	// записываем в лог, что сервер запускается
	sugar.Infow(
		"Starting server",
		"addr", config.Options.Address,
	)

	// Подгружаем URL из файла-хранилища в память-хранилище
	// При большой нагрузке это так себе решение, потому что съедим кучу оперативы, но для PoC - acceptable :)
	// В проде в любом случае необходимо использовать реальную СУБД, а не файлы/memory storage
	err = app.LoadURLsFromFile(config.Options.FileToWrite, sugar)
	if err != nil {
		return
	}
	err = http.ListenAndServe(config.Options.Address, router)
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
