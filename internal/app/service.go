package app

import (
	"database/sql"
	"github.com/JohnnyConstantin/urlshort/models"
	"sync"
)

var (
	mu sync.RWMutex
)

// Shortener интерфейс для разных функций бизнес-логики в зависимости от используемого StorageType
type Shortener interface {
	ShortenURL(originalURL string) models.ShortenResponse    // Используется для сжатия URL, используя оригинал
	GetFullURL(shortID string) (models.ShortenRequest, bool) // Используется для получения полного URL, используя короткий
}

func GetDBConnection(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return db, err
}
