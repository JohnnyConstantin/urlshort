package app

import (
	"github.com/JohnnyConstantin/urlshort/models"
	"github.com/google/uuid"
	"sync"
)

var (
	urlStore = make(map[string]string) // shortID: originalURL
	mu       sync.RWMutex
)

func shortenURL(originalURL string) models.ShortenResponse {
	var ShortenUrl models.ShortenResponse
	shortID := uuid.New().String()[:8]

	mu.Lock()
	urlStore[shortID] = originalURL
	mu.Unlock()

	ShortenUrl.Result = shortID

	return ShortenUrl
}

func getFullURL(shortID string) (models.URLResponse, bool) {
	var Result models.URLResponse
	mu.RLock()
	defer mu.RUnlock()
	originalURL, exists := urlStore[shortID]
	if exists {
		Result.OriginalURL = originalURL
		Result.ShortURL = shortID
		return Result, exists
	}
	return Result, exists
}
