package app

import (
	"database/sql"
	"sync"

	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
)

// DBFuller объект "разворачивания" URL c использованием БД
type DBFuller struct {
	DB  *sql.DB
	Cfg config.StorageConfig
}

// FileFuller объект "разворачивания" URL c использованием файла
type FileFuller struct {
	mu  *sync.Mutex
	Cfg config.StorageConfig
}

// MemoryFuller объект "разворачивания" URL c использованием хранилища в памяти
type MemoryFuller struct {
	mu  *sync.Mutex
	Cfg config.StorageConfig
}

// InitMutex создание мьютекса для файлового разворачивателя
func (f *FileFuller) InitMutex() {
	f.mu = new(sync.Mutex)
}

// InitMutex создание мьютекса для разворачивателя в памяти
func (f *MemoryFuller) InitMutex() {
	f.mu = new(sync.Mutex)
}

// InitMutex mock для БД
func (f *DBFuller) InitMutex() {
}

// GetFullURL получить из БД полную URL по сокращенному
func (f *DBFuller) GetFullURL(shortID string) (models.ShortenRequest, bool, bool, error) {
	result := models.ShortenRequest{URL: ""}

	originalURL, exists, isDeleted, err := store.Read(f.DB, shortID)
	if err != nil {
		return result, false, false, err
	}
	if exists {
		result.URL = originalURL
		return result, exists, isDeleted, nil
	}

	return result, exists, isDeleted, nil
}

// GetFullURL Получить из файла полную URL по сокращенной
func (f *FileFuller) GetFullURL(shortID string) (models.ShortenRequest, bool, bool, error) {
	result := models.ShortenRequest{URL: ""}

	f.mu.Lock()
	defer f.mu.Unlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		result.URL = originalURL
		return result, exists, false, nil
	}
	return result, exists, false, nil
}

// GetFullURL Получить из памяти полную URL по сокращенной
func (f *MemoryFuller) GetFullURL(shortID string) (models.ShortenRequest, bool, bool, error) {
	result := models.ShortenRequest{URL: ""}

	f.mu.Lock()
	defer f.mu.Unlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		result.URL = originalURL
		return result, exists, false, nil
	}
	return result, exists, false, nil
}
