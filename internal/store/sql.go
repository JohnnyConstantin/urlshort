package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/models"
)

func InitDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id          SERIAL PRIMARY KEY,
		short_url   VARCHAR(10) UNIQUE NOT NULL,
		original_url TEXT UNIQUE NOT NULL,
		created_at  TIMESTAMP DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);
	CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
	`
	_, err := db.ExecContext(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	return nil
}

func Insert(db *sql.DB, record models.URLRecord) (string, error) {
	shortKey := record.ShortURL
	originalURL := record.OriginalURL

	// Вставляем запись в БД (если originalURL уже есть, возвращаем существующий shortURL)
	var existingShortURL string
	err := db.QueryRow(`
		INSERT INTO urls (short_url, original_url)
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
		RETURNING short_url
	`, shortKey, originalURL).Scan(&existingShortURL)

	if err != nil {
		return "", fmt.Errorf("failed to insert URL: %w", err)
	}

	return existingShortURL, nil
}

func Read(db *sql.DB, shortID string) (string, bool) {
	var originalURL string

	err := db.QueryRow(
		`SELECT original_url FROM urls WHERE short_url = $1 LIMIT 1`,
		shortID,
	).Scan(&originalURL)

	switch {
	case err == nil:
		return originalURL, true
	case errors.Is(err, sql.ErrNoRows):
		return "", false
	default:
		return "", false
	}
}
