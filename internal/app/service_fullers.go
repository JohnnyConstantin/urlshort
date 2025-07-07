package app

import (
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
)

type DBFuller struct {
	cfg config.StorageConfig
}

type FileFuller struct {
	cfg config.StorageConfig
}

type MemoryFuller struct {
	cfg config.StorageConfig
}

func (f *DBFuller) GetFullURL(shortID string) (models.ShortenRequest, bool) {
	Result := models.ShortenRequest{URL: ""}
	fmt.Println("Fulling URL with DB: ", shortID)

	db, err := GetDBConnection(config.Options.DSN)
	defer db.Close()
	if err != nil {
		return Result, false
	}

	originalURL, exists := store.Read(db, shortID)
	if exists {
		Result.URL = originalURL
		return Result, exists
	}

	return Result, false
}

func (f *FileFuller) GetFullURL(shortID string) (models.ShortenRequest, bool) {
	Result := models.ShortenRequest{URL: ""}

	mu.RLock()
	defer mu.RUnlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		Result.URL = originalURL
		return Result, exists
	}
	return Result, exists
}

func (f *MemoryFuller) GetFullURL(shortID string) (models.ShortenRequest, bool) {
	Result := models.ShortenRequest{URL: ""}

	mu.RLock()
	defer mu.RUnlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		Result.URL = originalURL
		return Result, exists
	}
	return Result, exists
}
