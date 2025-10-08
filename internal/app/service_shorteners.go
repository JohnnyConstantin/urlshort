package app

import (
	"database/sql"
	"sync"

	"github.com/google/uuid"

	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
)

// DBShortener объект "сворачивания" URL c использованием БД
type DBShortener struct {
	db  *sql.DB
	cfg config.StorageConfig
}

// FileShortener объект "сворачивания" URL c использованием файла
type FileShortener struct {
	mu  *sync.Mutex
	cfg config.StorageConfig
}

// MemoryShortener объект "сворачивания" URL c использованием памяти
type MemoryShortener struct {
	mu  *sync.Mutex
	cfg config.StorageConfig
}

// InitMutex создание мьютекса для сворачивания с файлом
func (s *FileShortener) InitMutex() {
	s.mu = new(sync.Mutex)
}

// InitMutex создание мьютекса для сворачивания в памяти
func (s *MemoryShortener) InitMutex() {
	s.mu = new(sync.Mutex)
}

// ShortenURL сокращает URL с использованием БД
func (s *DBShortener) ShortenURL(userID string, originalURL string) (models.ShortenResponse, int) {
	var shortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	//Создаем объект для записи
	record := models.URLRecord{
		UUID:        uuid.New().String()[:4],
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	shortID, status, err := store.Insert(s.db, record, userID)
	if err != nil {
		return models.ShortenResponse{}, status
	}

	shortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return shortenURL, status
}

// ShortenURL сокращает URL с использованием файла
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

// ShortenURL сокращает URL с использованием памяти
func (s *MemoryShortener) ShortenURL(originalURL string) models.ShortenResponse {
	var shortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	s.mu.Lock()
	store.URLStore[shortID] = originalURL // сохраняем в память
	s.mu.Unlock()

	shortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return shortenURL
}
