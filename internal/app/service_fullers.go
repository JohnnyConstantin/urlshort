package app

import (
	"database/sql"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"sync"
)

type DBFuller struct {
	cfg config.StorageConfig
	db  *sql.DB
}

type FileFuller struct {
	cfg config.StorageConfig
	mu  *sync.Mutex
}

func (f *FileFuller) InitMutex() {
	f.mu = new(sync.Mutex)
}

type MemoryFuller struct {
	cfg config.StorageConfig
	mu  *sync.Mutex
}

func (f *MemoryFuller) InitMutex() {
	f.mu = new(sync.Mutex)
}

func (f *DBFuller) GetFullURL(shortID string, userID string) (models.ShortenRequest, bool) {
	result := models.ShortenRequest{URL: ""}
	fmt.Println("Fulling URL with DB: ", shortID)

	originalURL, exists, err := store.Read(f.db, shortID)
	if err != nil {
		return result, false
	}
	if exists {
		result.URL = originalURL
		return result, exists
	}

	return result, false
}

func (f *FileFuller) GetFullURL(shortID string) (models.ShortenRequest, bool) {
	result := models.ShortenRequest{URL: ""}

	f.mu.Lock()
	defer f.mu.Unlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		result.URL = originalURL
		return result, exists
	}
	return result, exists
}

func (f *MemoryFuller) GetFullURL(shortID string) (models.ShortenRequest, bool) {
	result := models.ShortenRequest{URL: ""}

	f.mu.Lock()
	defer f.mu.Unlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		result.URL = originalURL
		return result, exists
	}
	return result, exists
}
