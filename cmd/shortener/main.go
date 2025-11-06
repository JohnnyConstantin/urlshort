// Package main Осуществляет запуск сервера по сокращению ссылок, инициализирует конфигурацию и загружает переменные окружения
package main

import (
	"database/sql"
	"errors"
	"flag"
	"fmt"
	route "github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strings"

	"github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/certificates"
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

	//Чтобы удобнее было работать
	handler := server.Handler
	router := route.NewRouter() //Используем внешний роутер chi, вместо встроенного в объект app.Server

	//Создаём предустановленный регистратор zap
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer func(logger *zap.Logger) {
		logger.Sync()
	}(logger)

	sugar = *logger.Sugar() // Создали экземпляр и в дальнейшем прокидываем его в middleware с логированием

	flag.Parse()

	// Загружаем конфиг JSON. Логика перезаписывания с флагами инкапсулирована внутри. Переменные окружения и
	// так грузятся после этого
	config.LoadJSONConfig()

	// Вынес загрузку переменных окружения в отдельную функцию
	loadEnvs()

	// записываем в лог, что сервер запускается
	sugar.Infow(
		"Starting server",
		"addr", config.Options.Address,
	)

	// Создание и применение конфигурации. Если ошибка - выход с ненулевым кодом. Ошибка при этом логируется в sugar
	s.DB, err = storageDecider()
	if err != nil {
		panic(err)
	}

	// Запускаем HTTP-сервер для профилирования в отдельной горутине
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	// Вынес создание роутов в отдельную функцию
	createHandlers(s.DB, router, sugar, handler)

	startListenAndServe(s, router)
}

func startListenAndServe(s app.Server, router *route.Mux) {

	errChan := make(chan error, 1)         // Канал для ошибок. В данном контексте неважно из какой горутины прилетело - обработка идентичная для всех.
	shutdownChan := make(chan struct{}, 1) // Канал для graceful shutdown успешных сообщений

	if config.Options.EnableHTTPS {
		var cert, key = "cert.crt", "key.key" // Я бы сделал их также через опцию и env, но в задании об
		// этом явно не сказано, поэтому выбрал захардкодить, чтобы четко соответствовать ТЗ
		if !certificates.СertFilesExist(cert, key) { // Локально эти файлы есть, но в репу их не загружаю по понятным причинам
			err := certificates.GenerateCertAndPrivFiles(cert, key)
			if err != nil {
				sugar.Error("Failed to generate cert.crt/key.key") // Ошибка если не удалось создать пару
				return
			}
		}

		go func() {
			err := s.StartTLS(config.Options.Address, cert, key, router)
			if err != nil && !errors.Is(err, http.ErrServerClosed) { // Т.к. под капотом возвращается err в любом случае,
				// перехватываем только НЕ Shutdown\Close
				errChan <- err // Пробрасываем ошибку в канал
				return
			}
			errChan <- nil // Корректно завершились
		}()

		waitShutdown(errChan, shutdownChan, sugar, &s)

	} else {
		go func() {
			err := s.Start(config.Options.Address, router)
			if err != nil && !errors.Is(err, http.ErrServerClosed) { // Т.к. под капотом возвращается err в любом случае,
				// перехватываем только НЕ Shutdown\Close
				errChan <- err // Пробрасываем ошибку в канал
				return
			}
			errChan <- nil // Корректно завершились
		}()

		waitShutdown(errChan, shutdownChan, sugar, &s)
	}
}

// Функция ожидания сигнала завершения и обработки ошибок
func waitShutdown(errChan chan error, shutdownChan chan struct{}, sugar zap.SugaredLogger, s *app.Server) {

	go func() {
		err := s.WaitForShutdown(sugar) // Ожидаем сигнала
		if err != nil {
			sugar.Error("Failed while shutting down: ", err)
			errChan <- err
		}
		shutdownChan <- struct{}{} // Отправляем сигнал об успешном завершении
	}()

	// Ловим сигналы ошибок/успешного завершения
	select {
	case err := <-errChan: // Пришла ошибка
		if err != nil {
			sugar.Error("Received error while shutting down server", err)
			return
		}
		sugar.Info("Server gracefully shut down")
	case <-shutdownChan: // Пришел сигнал успешного завершения
		sugar.Info("Server gracefully shut down")
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
