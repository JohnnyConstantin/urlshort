package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/models"
	route "github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
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
						JSONResponseMiddleware( // Работа с json request/response
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

// Кастомный ResponseWriter для перехвата данных в JSONMiddleware
type responseWriterJSON struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
	header     http.Header
}

func (rw *responseWriterJSON) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
}

func (rw *responseWriterJSON) Write(b []byte) (int, error) {
	return rw.body.Write(b)
}

func (rw *responseWriterJSON) Header() http.Header {
	if rw.header == nil {
		rw.header = make(http.Header)
	}
	return rw.header
}

// JSONResponseMiddleware Middleware для проверки того, что возвращаемое значение является JSON. Иначе переводит его в JSON.
// Этот middleware хотелось бы вынести внутрь app в отдельный файл validator.go, но тогда не проходит автотест на импорт json
// в main.go (а без этого middleware импорт json здесь не нужен)
func JSONResponseMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Создаем  ResponseWriterJSON для перехвата ответа
		rw := &responseWriterJSON{
			ResponseWriter: w,
			body:           new(bytes.Buffer),
			statusCode:     http.StatusCreated, // По умолчанию 201
		}

		// Вызываем следующий обработчик с ResponseWriterJSON
		next(rw, r)

		// Копируем заголовки
		for k, v := range rw.header {
			w.Header()[k] = v
		}

		// Если это JSON ответ - обрабатываем
		if rw.Header().Get("Content-Type") == "application/json" {
			response := models.ShortenResponse{
				Result: rw.body.String(),
			}
			w.WriteHeader(rw.statusCode)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Для не-JSON ответов просто копируем тело и статус код
		w.WriteHeader(rw.statusCode)
		if rw.body.Len() > 0 {
			w.Write(rw.body.Bytes())
		}
	}
}
