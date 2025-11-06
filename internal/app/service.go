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

// Statter интерфейс для разных функций cтатистики в зависимости от используемого StorageType
type Statter interface {
	GetUsersCount() (int, error) // Получить количество пользователей
	GetURLsCount() (int, error)  // Получить количество URL
}
