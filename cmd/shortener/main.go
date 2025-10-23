// Package main Осуществляет запуск сервера по сокращению ссылок, инициализирует конфигурацию и загружает переменные окружения
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	route "github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
)

var sugar zap.SugaredLogger

// внутренние параметры для разработчика (предсказал их появление на первом же спринте xD). Перенес сюда из config.go
//
//nolint:gochecknoglobals
var (
	buildVersion = "N/A" // Версия билда
	buildDate    = "N/A" // Дата билда
	buildCommit  = "N/A" // Хеш коммита
)

func main() {

	// Вывод информации о сборке. Перетащил сюда глобальные переменные, чтобы их не экспортировать из другого пакета,
	// потому что изначально они находились пакете config
	printBuildInfo()

	var s app.Server
	server := s.NewServer()

	// Запускаем HTTP-сервер для профилирования в отдельной горутине
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	//Чтобы удобнее было работать
	handler := server.Handler
	router := route.NewRouter() //Используем внешний роутер chi, вместо встроенного в объект app.Server

	//Создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func(logger *zap.Logger) {
		err = logger.Sync()
		if err != nil {
			panic(err)
		}
	}(logger)

	sugar = *logger.Sugar() // Создали экземпляр и в дальнейшем прокидываем его в middleware с логированием

	flag.Parse()

	// Вынес загрузку переменных окружения в отдельную функцию
	loadEnvs()

	// записываем в лог, что сервер запускается
	sugar.Infow(
		"Starting server",
		"addr", config.Options.Address,
	)

	// Создание и применение конфигурации. Если ошибка - выход с ненулевым кодом. Ошибка при этом логируется в sugar
	db, err := storageDecider()
	if err != nil {
		panic(err)
	}

	//Если вернулся хендлер к БД (т.е. успешно создано соединение к БД), то закрываем после завершения программы
	if db != nil {
		defer func(db *sql.DB) {
			err = db.Close()
			if err != nil {
				panic(err)
			}
		}(db)
	}

	// Вынес создание роутов в отдельную функцию
	createHandlers(db, router, sugar, handler)

	if config.Options.EnableHTTPS {
		var cert, key = "cert.crt", "key.key" // Я бы сделал их также через опцию и env, но в задании об
		// этом явно не сказано, поэтому выбрал захардкодить, чтобы четко соответствовать ТЗ
		if !config.СertFilesExist(cert, key) { // Локально эти файлы есть, но в репу их не загружаю по понятным причинам
			err = config.GenerateCertAndPrivFiles(cert, key)
			if err != nil {
				sugar.Error("Failed to generate cert.crt/key.key") // Ошибка если не удалось создать пару
				return
			}
		}
		err = http.ListenAndServeTLS(config.Options.Address, cert, key, router)
	} else {
		err = http.ListenAndServe(config.Options.Address, router)
	}

	if err != nil {
		return
	}

}

func createHandlers(db *sql.DB, router *route.Mux, sugar zap.SugaredLogger, handler *app.Handler) {
	//Накидываем хендлеры на роуты
	router.Route("/", func(r route.Router) {
		r.Post("/",
			app.GzipHandle( // Сжатие
				app.WithLogging(db,
					handler.WithAuth( // Логирование, прокидываем в него регистратор логов sugar
						handler.PostHandler), sugar))) // Сам хендлер
		r.Route("/api", func(r route.Router) {
			r.Route("/shorten", func(r route.Router) {
				r.Post("/",
					app.GzipHandle( // Сжатие
						app.WithLogging(db, // Логирование, прокидываем в него регистратор логов sugar
							handler.WithAuth( // Добавляем аутентификацию
								handler.PostHandler), sugar))) // Сам хендлер
				r.Post("/batch",
					app.GzipHandle( // Сжатие
						app.WithLogging(db,
							handler.WithAuth( // Логирование, прокидываем в него регистратор логов sugar
								handler.PostHandlerMultiple), sugar))) // Сам хендлер

			})
			r.Route("/user", func(r route.Router) {
				r.Delete("/urls",
					app.GzipHandle( // Сжатие
						app.WithLogging(db, // Логирование, прокидываем в него регистратор логов sugar
							handler.WithAuth( // Добавляем аутентификацию
								handler.DeleteHandlerMultiple), sugar))) // Сам хендлер
				r.Get(
					"/urls",
					app.GzipHandle( // Сжатие
						app.WithLogging(db, // Логирование, прокидываем в него регистратор логов sugar
							handler.WithAuth( //Добавляем аутентификацию
								handler.GetHandlerMultiple), sugar))) // Сам хендлер

			})
		})
		r.Get("/{id}",
			app.GzipHandle( // Сжатие
				app.WithLogging(db, // Логирование, прокидываем в него регистратор логов sugar
					handler.GetHandler, sugar))) // Сам хендлер
		r.Get("/ping",
			app.WithLogging(db,
				handler.PingDBHandler, sugar)) // Сам хендлер
	})
}

func loadEnvs() {
	//Подгружаем переменные окружения при наличии
	envA, ok := os.LookupEnv("SERVER_ADDRESS")
	if ok && envA != "" {
		config.Options.Address = envA
	}
	envB, ok := os.LookupEnv("BASE_URL")
	if ok && envB != "" {
		config.Options.BaseAddress = envB
	}
	envC, ok := os.LookupEnv("FILE_STORAGE_PATH")
	if ok && envC != "" {
		config.Options.FileToWrite = envC
	}

	envD, ok := os.LookupEnv("DATABASE_DSN")
	if ok && envD != "" {
		config.Options.DSN = envD
	}

	envE, ok := os.LookupEnv("SECRET_KEY")
	if ok && envE != "" {
		config.Options.SecretKey = envE
	}

	envF, ok := os.LookupEnv("ENABLE_HTTPS")
	if ok && (strings.ToLower(envF) == "true" || strings.ToLower(envF) == "1") { // Предполагаю, что переменная
		// окружения должна содержать true или 1 (перевожу в lowercase, чтобы обработать True и TRUE)
		config.Options.EnableHTTPS = true
	}

}

func storageDecider() (*sql.DB, error) {
	// Вызываем резолвер способа хранения данных
	config.CreateStorageConfig()
	cfg := config.GetStorageConfig()

	//Логируем какой StorageType будет использован, для FileMemory выполняем операцию по восстановлению из файла
	switch cfg.StorageType {
	case config.StorageDB:
		db := store.DB{}
		err := db.OpenDB(config.Options.DSN)
		if err != nil {
			sugar.Error("Could not connect to database")
			return nil, err
		}

		// Создаем таблицу (если ее нет)
		if err = db.InitDB(); err != nil {
			sugar.Error("Could not initialize database")
			return nil, err
		}

		sugar.Infow("Using PostgreSQL as a storage",
			"DSN", config.Options.DSN)

		return db.DB, nil

	case config.StorageFile:
		sugar.Infow("Using file as a storage",
			"file", config.Options.FileToWrite)

		// Подгружаем URL из файла-хранилища в память-хранилище
		// При большой нагрузке это так себе решение, потому что съедим кучу оперативы, но для PoC - acceptable :)
		// Вариант с "постоянно дергать файл на Read/Write операции" без использования in-Memory показался совсем варварским
		err := app.LoadURLsFromFile(config.Options.FileToWrite, sugar)
		if err != nil {
			return nil, err
		}
	default:
		//Только логируем, никаких доп.действий не требуется, все реализовано через проверку StorageType в целевых функциях
		sugar.Infow("Using memory storage (no persistence)")
	}
	return nil, nil
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
