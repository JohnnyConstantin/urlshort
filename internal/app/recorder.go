package app

import (
	"bufio"
	"encoding/json"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"go.uber.org/zap"
	"io"
	"os"
)

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

	// Проверяем существование файла перед открытием
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			logger.Infof("Файл %s не существует, пропускаем загрузку URL", filename)
			return nil
		}
		return nil
	}

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
		store.URLStore[record.ShortURL] = record.OriginalURL

		logger.Infoln("Added to memory: " + store.URLStore[record.ShortURL])
	}

	return nil
}
