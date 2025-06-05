package app

import (
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"github.com/JohnnyConstantin/urlshort/models"
	"github.com/google/uuid"
	"sync"
)

var (
	mu sync.RWMutex
)

func shortenURL(originalURL string) models.ShortenResponse {
	var ShortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	mu.Lock()
	store.UrlStore[shortID] = originalURL
	mu.Unlock()

	ShortenURL.Result = shortID

	return ShortenURL
}

func getFullURL(shortID string) (models.URLResponse, bool) {
	var Result models.URLResponse

	mu.RLock()
	defer mu.RUnlock()
	originalURL, exists := store.UrlStore[shortID]
	if exists {
		Result.OriginalURL = originalURL
		Result.ShortURL = shortID
		return Result, exists
	}
	return Result, exists
}
