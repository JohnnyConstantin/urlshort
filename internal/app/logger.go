package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type (
	key    string
	logger string
)

const (
	dbKey     key    = "database"
	loggerKey logger = "sugar"
)

type (
	// Берём структуру для хранения сведений об ответе
	responseData struct {
		status int
		size   int
	}

	// Добавляем реализацию http.ResponseWriter
	loggingResponseWriter struct {
		http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
		responseData        *responseData
	}
)

// Реализация метода Write для logger
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// Записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

// Реализация метода WriteHeader для logger
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// Записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

// WithLogging мидлварь, отвечающая за логирование запросов и ответов
func WithLogging(db *sql.DB, h http.HandlerFunc, logger zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Засекаем

		ctx := context.WithValue(r.Context(), dbKey, db)
		ctx = context.WithValue(ctx, loggerKey, logger)

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		// Прокидываем дальше
		h(&lw, r.WithContext(ctx))

		duration := time.Since(start) // Получаем время выполнения всех последующих middleware хендлеров
		//todo Кажется, лучше вынести засекание времени выполнения в отдельный middleware с хранением времени старта.
		// Иначе получается, что засекание времени происходит в конвейере middleware в том месте, где очередь доходит
		// до этого хендлера. А все, что было до него, не учитывается в финальном подсчете. Как итог: неактуальный diration
		// но в рамках текущего ТЗ это не так важно :)

		logger.Infoln( // Логируем
			"uri", r.RequestURI,
			"method", r.Method,
			"status", responseData.status,
			"duration", duration,
			"size", responseData.size,
		)
	}
}
