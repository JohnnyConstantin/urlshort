package app

import (
	"database/sql"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"sync"
)

var (
	mu sync.RWMutex
)

func shortenURL(db *sql.DB, originalURL string) models.ShortenResponse {
	var ShortenURL models.ShortenResponse

	shortURL, exists := store.IsDuplicate(db, originalURL)
	if exists {
		ShortenURL.Result = shortURL
		return ShortenURL
	}

	shortID, err := store.Insert(db, originalURL)
	if err != nil {
		return ShortenURL
	}

	ShortenURL.Result = shortID

	return ShortenURL
}

func getFullURL(db *sql.DB, shortID string) (models.URLResponse, bool) {
	var Result models.URLResponse

	originalURL, err := store.Read(db, shortID)
	if err != nil {
		return Result, false
	}

	Result.OriginalURL = originalURL
	Result.ShortURL = shortID
	return Result, true
}
