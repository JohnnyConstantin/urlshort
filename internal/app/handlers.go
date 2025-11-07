package app

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"io"
	"net/http"
	_ "net/http/pprof"
	"strings"

	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
)

// Handler Объект хендлера
type Handler struct {
	router  *Router
	Service *Service
}

// NewHandler Инциализация объекта хендлера с пустым роутером
func NewHandler(service *Service) *Handler {
	h := &Handler{
		router:  NewRouter(),
		Service: service,
	}

	return h
}

// ServeHTTP Утиная типизация, прокидываемся до функциональной части роутера по роутингу запросов на хендлер
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

// GetHandler обрабатывает GET запросы
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	response := models.ShortenRequest{URL: ""}
	exists := false //By default не существует
	var isDeleted bool
	var status int
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")

	ctx := r.Context()

	sugar, ok := ctx.Value(loggerKey).(zap.SugaredLogger)
	if !ok {
		http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
		return
	}

	if len(parts) != 1 {
		http.Error(w, store.BadRequestError, store.DefaultErrorCode)
		return
	}

	id := parts[0]

	cfg := config.GetStorageConfig()

	status = http.StatusTemporaryRedirect //default

	switch cfg.StorageType {
	case config.StorageFile:
		h.Service.Fuller = &FileFuller{Cfg: cfg}
		h.Service.Fuller.InitMutex()
		response, exists, _ = h.Service.Fuller.GetFullURL(id)
	case config.StorageMemory:
		h.Service.Fuller = &MemoryFuller{Cfg: cfg}
		h.Service.Fuller.InitMutex()
		response, exists, _ = h.Service.Fuller.GetFullURL(id)
	case config.StorageDB:
		// Если StorageDB, то в context не может быть nil (на это есть проверка в main), однако, на всякий случай здесь повторяем
		db, ok := r.Context().Value(DbKey).(*sql.DB)
		if !ok {
			sugar.Errorf("Not supported storage type: %v", cfg.StorageType)
			http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
			return
		}

		h.Service.Fuller = &DBFuller{db, cfg}
		response, exists, isDeleted = h.Service.Fuller.GetFullURL(id)
		if isDeleted {
			status = http.StatusGone
		}
	default:
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	if !exists {
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	result := response.URL

	w.Header().Set("Location", result)

	w.WriteHeader(status)
}

// PostHandler обрабатывает POST запросы
func (h *Handler) PostHandler(w http.ResponseWriter, r *http.Request) {
	var ShortURL models.ShortenResponse
	var OriginalURL models.ShortenRequest
	var status = 201

	ctx := r.Context()

	sugar, ok := ctx.Value(loggerKey).(zap.SugaredLogger)
	if !ok {
		http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
		return
	}

	// Доп. обработка тела
	// Ограничиваем размер тела запроса
	maxSize := int64(1024 * 1024)
	limitedReader := io.LimitReader(r.Body, maxSize)

	body, err := io.ReadAll(limitedReader)
	if err != nil {
		sugar.Errorf("Error in reading request body: %v", err)
		http.Error(w, store.ReadBodyError, store.DefaultErrorCode)
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(r.Body)

	if r.Header.Get("Content-Type") == "application/json" {
		if err = json.Unmarshal(body, &OriginalURL); err != nil {
			http.Error(w, store.DefaultError, store.DefaultErrorCode)
		}
		w.Header().Set("Content-Type", "application/json")
	} else {
		OriginalURL.URL = string(body)
		w.Header().Set("Content-Type", "text/plain")
	}

	cfg := config.GetStorageConfig()
	switch cfg.StorageType {
	case config.StorageFile:
		h.Service.Shortener = &FileShortener{Cfg: cfg}
		h.Service.Shortener.InitMutex()
		status = http.StatusCreated
		shorten_req := Shortenerequest{OriginalURL: OriginalURL.URL}
		ShortURL = h.Service.Shortener.ShortenURL(shorten_req)
	case config.StorageMemory:
		h.Service.Shortener = &MemoryShortener{Cfg: cfg}
		h.Service.Shortener.InitMutex()
		status = http.StatusCreated
		shorten_req := Shortenerequest{OriginalURL: OriginalURL.URL}
		ShortURL = h.Service.Shortener.ShortenURL(shorten_req)
	case config.StorageDB:
		db, userID, errs := initCtx(r)
		if errs != nil {
			sugar.Errorf("Error in initialization of db and userID: %v", errs)
			http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
			return
		}
		h.Service.Shortener = &DBShortener{db, cfg}
		shorten_req := Shortenerequest{OriginalURL: OriginalURL.URL, UserID: userID}
		ShortURL = h.Service.Shortener.ShortenURL(shorten_req)
	default:
		sugar.Errorf("Unsupported storage type: %v", cfg.StorageType)
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	// Перенесенный функционал из JsonMiddleware. Необходимо для применения статус кода и json encoding для app/json Header
	if r.Header.Get("Content-Type") == "application/json" {
		w.WriteHeader(status)
		err = json.NewEncoder(w).Encode(ShortURL)
		if err != nil {
			sugar.Errorf("Error in encoding response body: %v", err)
			// Добавил логирование, сменил статус на InternalErrorCode, текст оставил дефолтным (он по тз).
			// Также прокинул во все хендлеры логгер и логирую ошибки типа InternalError
			http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
			return
		}
	} else {
		w.WriteHeader(status)
		_, err := w.Write([]byte(ShortURL.Result))
		if err != nil {
			sugar.Errorf("Error in writing response body: %v", err)
			http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
			return
		}
	}
}

// PostHandlerMultiple обрабатывает POST запросы с batch
func (h *Handler) PostHandlerMultiple(w http.ResponseWriter, r *http.Request) {
	var requests []models.BatchShortenRequest
	var ShortURL models.ShortenResponse
	status := http.StatusCreated

	ctx := r.Context()

	sugar, ok := ctx.Value(loggerKey).(zap.SugaredLogger)
	if !ok {
		http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
		return
	}

	// Доп. обработка тела
	// Ограничиваем размер тела запроса
	maxSize := int64(1024 * 1024)
	limitedReader := io.LimitReader(r.Body, maxSize)
	body, err := io.ReadAll(limitedReader)

	if err != nil {
		http.Error(w, store.ReadBodyError, store.DefaultErrorCode)
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			return
		}
	}(r.Body)

	if err := json.Unmarshal(body, &requests); err != nil {
		http.Error(w, "Invalid batch request format", store.DefaultErrorCode)
		return
	}

	responses := make([]models.BatchShortenResponse, 0, len(requests))
	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		h.Service.Shortener = &FileShortener{Cfg: cfg}
		h.Service.Shortener.InitMutex()
		for _, req := range requests {
			shorten_req := Shortenerequest{OriginalURL: req.OriginalURL}
			ShortURL = h.Service.Shortener.ShortenURL(shorten_req)

			responses = append(responses, models.BatchShortenResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      ShortURL.Result,
			})
		}

	case config.StorageMemory:
		h.Service.Shortener = &MemoryShortener{Cfg: cfg}
		h.Service.Shortener.InitMutex()
		for _, req := range requests {
			shorten_req := Shortenerequest{OriginalURL: req.OriginalURL}
			ShortURL = h.Service.Shortener.ShortenURL(shorten_req)

			responses = append(responses, models.BatchShortenResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      ShortURL.Result,
			})
		}
	case config.StorageDB:
		db, userID, err := initCtx(r)
		if err != nil {
			sugar.Errorf("Error in initialization of db and userID: %v", err)
			http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
			return
		}
		h.Service.Shortener = &DBShortener{db, cfg}
		for _, req := range requests {
			shorten_req := Shortenerequest{OriginalURL: req.OriginalURL, UserID: userID}
			ShortURL = h.Service.Shortener.ShortenURL(shorten_req)

			responses = append(responses, models.BatchShortenResponse{
				CorrelationID: req.CorrelationID,
				ShortURL:      ShortURL.Result,
			})
		}
	default: // Overkill, но перестраховаться нужно
		sugar.Errorf("Unsupported storage type: %v", cfg.StorageType)
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(responses); err != nil {
		sugar.Errorf("Error in encoding response body: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// DeleteHandlerMultiple удаляет несколько записей ссылок
func (h *Handler) DeleteHandlerMultiple(w http.ResponseWriter, r *http.Request) {
	db, userID, err := initCtx(r)
	if err != nil {
		http.Error(w, err.Error(), store.InternalSeverErrorCode)
		return
	}

	ctx := r.Context()

	sugar, ok := ctx.Value(loggerKey).(zap.SugaredLogger)
	if !ok {
		http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
		return
	}

	deleter := DBDeleter{cfg: config.GetStorageConfig(), db: db}

	// Парсим тело запроса
	var shortURLs []string
	if err := json.NewDecoder(r.Body).Decode(&shortURLs); err != nil {
		sugar.Errorf("Error in decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := deleter.DeleteURL(userID, shortURLs); err != nil {
		sugar.Errorf("Error in deleting URL: %v", err)
		http.Error(w, store.DefaultError, http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusAccepted)
}

// GetHandlerMultiple получить несколько полных URL по сокращенным
func (h *Handler) GetHandlerMultiple(w http.ResponseWriter, r *http.Request) {
	db, userID, err := initCtx(r)
	if err != nil {
		http.Error(w, err.Error(), store.InternalSeverErrorCode)
		return
	}

	ctx := r.Context()

	sugar, ok := ctx.Value(loggerKey).(zap.SugaredLogger)
	if !ok {
		http.Error(w, store.DefaultError, store.InternalSeverErrorCode)
		return
	}

	urls, err := store.ReadWithUUID(db, userID)
	if err != nil {
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	if len(urls) == 0 {
		http.Error(w, store.DefaultError, http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(urls); err != nil {
		sugar.Errorf("Error in encoding response body: %v", err)
		http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (h *Handler) GetHandlerStats(w http.ResponseWriter, r *http.Request) {
	var stats models.Statistics
	var statistics Statter

	ctx := r.Context()

	// Извлекаем логгер из контекста
	sugar, ok := ctx.Value(loggerKey).(zap.SugaredLogger)
	if !ok {
		return
	}

	// Извлекаем бд из контекста
	db, ok := r.Context().Value(DbKey).(*sql.DB)
	if !ok {
		sugar.Errorf("Error in getting DB connection from context")
		return
	}

	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		statistics = NewFileStatistics(cfg)

	case config.StorageMemory:
		statistics = NewMemoryStatistics(cfg)

	case config.StorageDB:
		statistics = NewDBStatistics(db, cfg)

	default: // Overkill, но перестраховаться нужно
		sugar.Errorf("Unsupported storage type: %v", cfg.StorageType)
		http.Error(w, store.DefaultError, store.DefaultErrorCode)
		return
	}

	cnt, err := statistics.GetURLsCount()
	if err != nil {
		return
	}
	usrs, err := statistics.GetUsersCount()
	if err != nil {
		return
	}

	stats.UsersCount = usrs
	stats.URLCount = cnt

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		sugar.Errorf("Error in encoding response body: %v", err)
		http.Error(w, "JSON encoding failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GzipHandle мидлварь для работы со сжатием
func GzipHandle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверка на то, что клиент прислал пожатый контент
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body) // Распаковываем
			if err != nil {
				http.Error(w, "Invalid gzip body", http.StatusBadRequest)
				return
			}
			defer func(gz *gzip.Reader) {
				err = gz.Close()
				if err != nil {
					return
				}
			}(gz)
			r.Body = gz
		}

		originalWriter := w
		acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") // Проверяем что клиент поддерживает сжатие

		if acceptsGzip { // Если клиент поддерживает сжатие, проверяем передаваемый content-type
			contentType := r.Header.Get("Content-Type")
			if strings.HasPrefix(contentType, "application/json") ||
				strings.HasPrefix(contentType, "text/html") {

				gzWriter := gzip.NewWriter(w) // Жмем!
				defer func(gzWriter *gzip.Writer) {
					err := gzWriter.Close()
					if err != nil {
						return
					}
				}(gzWriter)

				w.Header().Set("Content-Encoding", "gzip") // Ставим заголовок, что пожали контент
				originalWriter = &gzipWriter{
					ResponseWriter: w,
					Writer:         gzWriter,
				}
			}
		}

		next(originalWriter, r) // перекидываем дальше
	}
}

// PingDBHandler Проверяет подключение к БД
func (h *Handler) PingDBHandler(w http.ResponseWriter, r *http.Request) {
	cfg := config.GetStorageConfig()
	if cfg.StorageType == config.StorageDB {
		db, ok := r.Context().Value(DbKey).(*sql.DB)
		if !ok {
			http.Error(w, store.ConnectionError, http.StatusInternalServerError)
		}
		err := db.Ping()
		if err != nil {
			http.Error(w, store.ConnectionError, store.InternalSeverErrorCode)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func initCtx(r *http.Request) (*sql.DB, string, error) {
	db, ok := r.Context().Value(DbKey).(*sql.DB)
	if !ok {
		return nil, "", errors.New("DB not in context")
	}
	userID, ok := r.Context().Value(User).(string)
	if !ok {
		return nil, "", errors.New("userID not found in context")
	}

	return db, userID, nil
}
