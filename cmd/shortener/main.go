package main

import (
	"flag"
	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
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
			r.Route("/shorten", func(r route.Router) {
				r.Post("/",
					app.GzipHandle( // Сжатие
						app.WithLogging( // Логирование, прокидываем в него регистратор логов sugar
							handler.PostHandler, sugar))) // Сам хендлер
				r.Post("/batch",
					app.GzipHandle( // Сжатие
						app.WithLogging( // Логирование, прокидываем в него регистратор логов sugar
							handler.PostHandlerMultiple, sugar))) // Сам хендлер

			})
		})
		r.Get("/{id}",
			app.GzipHandle( // Сжатие
				app.WithLogging( // Логирование, прокидываем в него регистратор логов sugar
					handler.GetHandler, sugar))) // Сам хендлер
		r.Get("/ping",
			handler.PingDBHandler) // Сам хендлер
	})

	flag.Parse()

	// Вынес загрузку переменных окружения в отдельную функцию
	loadEnvs()

	// записываем в лог, что сервер запускается
	sugar.Infow(
		"Starting server",
		"addr", config.Options.Address,
	)

	// Вызываем резолвер способа хранения данных
	config.CreateStorageConfig()
	cfg := config.GetStorageConfig()

	//Логируем какой StorageType будет использован, для in-memory выполняем операцию по восстановлению из файла
	switch cfg.StorageType {
	case config.StorageDB:
		// Проверяем насколько верный DSN
		db, err := app.GetDBConnection(config.Options.DSN)
		if err != nil {
			sugar.Error("Could not connect to database")
			return
		}
		defer db.Close()

		// Создаем таблицу (если ее нет)
		if err := store.InitDB(db); err != nil {
			sugar.Error("Could not initialize database")
			return
		}

		sugar.Infow("Using PostgreSQL as a storage",
			"DSN", config.Options.DSN)

	case config.StorageFile:
		sugar.Infow("Using file as a storage",
			"file", config.Options.FileToWrite)

		// Подгружаем URL из файла-хранилища в память-хранилище
		// При большой нагрузке это так себе решение, потому что съедим кучу оперативы, но для PoC - acceptable :)
		// Вариант с "постоянно дергать файл на Read/Write операции" без использования in-Memory показался совсем варварским
		err = app.LoadURLsFromFile(config.Options.FileToWrite, sugar)
		if err != nil {
			return
		}
	default:
		//Только логируем, никаких доп.действий не требуется, все реализовано через проверку StorageType в целевых функциях
		sugar.Infow("Using memory storage (no persistence)")
	}

	err = http.ListenAndServe(config.Options.Address, router)
	if err != nil {
		return
	}
}

func loadEnvs() {
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

	if envD := os.Getenv("DATABASE_DSN"); envD != "" {
		config.Options.DSN = envD
	}
}
