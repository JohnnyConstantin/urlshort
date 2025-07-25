package app

import (
	"database/sql"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"github.com/google/uuid"
	"sync"
)

type DBShortener struct {
	cfg config.StorageConfig
	db  *sql.DB
}

type FileShortener struct {
	cfg config.StorageConfig
	mu  *sync.Mutex
}

func (s *FileShortener) InitMutex() {
	s.mu = new(sync.Mutex)
}

type MemoryShortener struct {
	cfg config.StorageConfig
	mu  *sync.Mutex
}

func (s *MemoryShortener) InitMutex() {
	s.mu = new(sync.Mutex)
}

func (s *DBShortener) ShortenURL(originalURL string) (models.ShortenResponse, int) {
	var shortenURL models.ShortenResponse

	fmt.Println("Shortening URL for DB: ", originalURL)

	shortID := uuid.New().String()[:8]

	//Создаем объект для записи
	record := models.URLRecord{
		UUID:        uuid.New().String()[:4],
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	shortID, status, err := store.Insert(s.db, record)
	if err != nil {
		return models.ShortenResponse{}, store.InternalSeverErrorCode
	}

	shortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return shortenURL, status
}

func (s *FileShortener) ShortenURL(originalURL string) models.ShortenResponse {
	var shortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	//Создаем объект для записи
	record := models.URLRecord{
		UUID:        uuid.New().String()[:4],
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	s.mu.Lock()
	store.URLStore[shortID] = originalURL // сохраняем в память
	s.mu.Unlock()
	err := SaveToFile(record) // сохраняем в файл
	if err != nil {
		return models.ShortenResponse{}
	}

	shortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return shortenURL
}

func (s *MemoryShortener) ShortenURL(originalURL string) models.ShortenResponse {
	var shortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	s.mu.Lock()
	store.URLStore[shortID] = originalURL // сохраняем в память
	s.mu.Unlock()

	shortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return shortenURL
}
