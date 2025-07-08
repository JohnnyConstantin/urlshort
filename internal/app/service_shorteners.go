package app

import (
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"github.com/google/uuid"
	"sync"
)

type DBShortener struct {
	cfg config.StorageConfig
	mu  sync.RWMutex
}

type FileShortener struct {
	cfg config.StorageConfig
	mu  sync.RWMutex
}

type MemoryShortener struct {
	cfg config.StorageConfig
	mu  sync.RWMutex
}

func (s *DBShortener) ShortenURL(originalURL string) (models.ShortenResponse, int) {
	var ShortenURL models.ShortenResponse

	fmt.Println("Shortening URL for DB: ", originalURL)

	shortID := uuid.New().String()[:8]

	//Создаем объект для записи
	record := models.URLRecord{
		UUID:        uuid.New().String()[:4],
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	db, err := GetDBConnection(config.Options.DSN)

	if err != nil {
		return models.ShortenResponse{}, store.InternalSeverErrorCode
	}
	defer db.Close()

	shortID, status, err := store.Insert(db, record)
	if err != nil {
		return models.ShortenResponse{}, store.InternalSeverErrorCode
	}

	ShortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return ShortenURL, status
}

func (s *FileShortener) ShortenURL(originalURL string) models.ShortenResponse {
	var ShortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	//Создаем объект для записи
	record := models.URLRecord{
		UUID:        uuid.New().String()[:4],
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	s.mu.Lock()
	store.URLStore[shortID] = originalURL // сохраняем в память
	err := SaveToFile(record)             // сохраняем в файл
	if err != nil {
		return models.ShortenResponse{}
	}
	s.mu.Unlock()

	ShortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return ShortenURL
}

func (s *MemoryShortener) ShortenURL(originalURL string) models.ShortenResponse {
	var ShortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	s.mu.Lock()
	store.URLStore[shortID] = originalURL // сохраняем в память
	s.mu.Unlock()

	ShortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return ShortenURL
}
