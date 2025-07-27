package app

import (
	"github.com/JohnnyConstantin/urlshort/models"
)

// Shortener интерфейс для разных функций бизнес-логики в зависимости от используемого StorageType
type Shortener interface {
	ShortenURL(originalURL string) models.ShortenResponse    // Используется для сжатия URL, используя оригинал
	GetFullURL(shortID string) (models.ShortenRequest, bool) // Используется для получения полного URL, используя короткий
	DeleteURL(userID string) error
}
