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

// Используется для получения короткого URL, используя полный
func shortenURL(originalURL string) models.ShortenResponse {
	var ShortenURL models.ShortenResponse

	shortID := uuid.New().String()[:8]

	//Оверкилл, но в будущем может пригодиться при использовании горутин на хендлерах
	mu.Lock()
	store.URLStore[shortID] = originalURL
	mu.Unlock()

	ShortenURL.Result = shortID

	return ShortenURL
}

// Используется для получения полного URL, используя короткий
func getFullURL(shortID string) (models.URLResponse, bool) {
	var Result models.URLResponse

	//Оверкилл, но в будущем может пригодиться при использовании горутин на хендлерах
	mu.RLock()
	defer mu.RUnlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		Result.OriginalURL = originalURL
		Result.ShortURL = shortID
		return Result, exists
	}
	return Result, exists
}
