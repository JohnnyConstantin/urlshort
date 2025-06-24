package app

import (
	"bufio"
	"encoding/json"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"go.uber.org/zap"
	"io"
	"net/http"
	"os"
	"time"
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

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// Записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// Записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func WithLogging(h http.HandlerFunc, logger zap.SugaredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now() // Засекаем

		responseData := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}
		// Прокидываем дальше
		h(&lw, r)

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

// SaveToFile сохранение объекта URLRecord в файл
func SaveToFile(event models.URLRecord) error {
	file, err := os.OpenFile(config.Options.FileToWrite, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(&event)
	if err != nil {
		return err
	}

	writer := bufio.NewWriter(file)

	if _, err := writer.Write(data); err != nil {
		return err
	}

	if err := writer.WriteByte('\n'); err != nil {
		return err
	}

	return writer.Flush()
}

// LoadURLsFromFile загрузка записей из файла в память
func LoadURLsFromFile(filename string, logger zap.SugaredLogger) error {
	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	for {
		var record models.URLRecord
		if err := decoder.Decode(&record); err != nil {
			if err == io.EOF {
				break
			}
			logger.Error("Ошибка декодирования JSON при чтении из файлового хранилища: %v", err)
			continue // Пропускаем некорректные записи (но логируем их)
		}

		// Записываем в память
		mu.Lock()
		store.URLStore[record.ShortURL] = record.OriginalURL

		logger.Infoln("Added to memory: " + store.URLStore[record.ShortURL])
		mu.Unlock()
	}

	return nil
}
