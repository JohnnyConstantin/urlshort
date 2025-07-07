package app

import (
	"github.com/JohnnyConstantin/urlshort/internal/config"
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

	//Создаем объект для записи в файл
	record := models.URLRecord{
		UUID:        uuid.New().String()[:4],
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}

	//Оверкилл, но в будущем может пригодиться при использовании горутин на хендлерах
	mu.Lock()
	store.URLStore[shortID] = originalURL // сохраняем в память
	err := SaveToFile(record)             // сохраняем в файл
	if err != nil {
		return models.ShortenResponse{}
	}
	mu.Unlock()

	ShortenURL.Result = config.Options.BaseAddress + "/" + shortID

	return ShortenURL
}

// Используется для получения полного URL, используя короткий
func getFullURL(shortID string) (models.ShortenRequest, bool) {
	Result := models.ShortenRequest{URL: ""}

	//Оверкилл, но в будущем может пригодиться при использовании горутин на хендлерах
	mu.RLock()
	defer mu.RUnlock()
	originalURL, exists := store.URLStore[shortID]
	if exists {
		Result.URL = originalURL
		return Result, exists
	}
	return Result, exists
}
