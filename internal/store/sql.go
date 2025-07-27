package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/models"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"net/http"
)

type DB struct {
	DB *sql.DB
}

func (d *DB) OpenDB(connStr string) error {
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к БД: %v", err)
	}

	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("не удалось проверить подключение: %v", err)
	}

	d.DB = sqlDB
	return nil
}

// InitDB создает базу данных, если ее нет
func (d *DB) InitDB() error {
	query := `
    CREATE TABLE IF NOT EXISTS urls (
        id          SERIAL PRIMARY KEY,
        uuid        VARCHAR(36),
        short_url   VARCHAR(10) UNIQUE NOT NULL,
        original_url TEXT UNIQUE NOT NULL,
        is_deleted  BOOLEAN DEFAULT FALSE,
        created_at  TIMESTAMP DEFAULT NOW()
    );
    CREATE INDEX IF NOT EXISTS idx_short_url ON urls(short_url);
    CREATE INDEX IF NOT EXISTS idx_original_url ON urls(original_url);
    CREATE INDEX IF NOT EXISTS idx_is_deleted ON urls(is_deleted);
    `
	_, err := d.DB.ExecContext(context.Background(), query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	return nil
}

// Insert вставляет originalURL и shortKey в БД
func Insert(db *sql.DB, record models.URLRecord, uuid string) (string, int, error) {
	var existingShortURL string
	var status int
	shortKey := record.ShortURL
	originalURL := record.OriginalURL

	// Вставляем запись в БД (если OriginalURL уже есть, возвращаем существующий shortURL)
	err := db.QueryRow(`
        WITH insert_attempt AS (
            INSERT INTO urls (uuid, short_url, original_url)
            VALUES ($1, $2, $3)
            ON CONFLICT (original_url) DO NOTHING
            RETURNING short_url
        )
        SELECT * FROM insert_attempt
        UNION
        SELECT short_url FROM urls WHERE original_url = $3 AND is_deleted = false
        LIMIT 1
    `, uuid, shortKey, originalURL).Scan(&existingShortURL)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Товарищи в пачке уже сталкивались, у UniqueViolation именно такой код
			// Просто возвращаем ошибку. Можно было бы перегенерировать shortURL, либо сделать роллбек через транзакции, но в ТЗ этого не указано, поэтому просто Error
			return existingShortURL, http.StatusConflict, nil
		}
		return existingShortURL, http.StatusConflict, fmt.Errorf("database error: %v", err)
	}

	if existingShortURL == record.ShortURL {
		status = http.StatusCreated
	} else {
		status = http.StatusConflict
	}

	return existingShortURL, status, nil
}

// Read Вычитывает original_url по shortID
func Read(db *sql.DB, shortID string) (string, bool, error) {
	var originalURL string

	err := db.QueryRow(
		`SELECT original_url FROM urls WHERE short_url = $1 LIMIT 1`,
		shortID,
	).Scan(&originalURL)

	switch {
	case err == nil:
		return originalURL, true, nil
	case errors.Is(err, sql.ErrNoRows):
		return "", false, err
	default:
		return "", false, err
	}
}

// ReadWithUUID Вычитывает original_url по shortID и userID
func ReadWithUUID(db *sql.DB, userID string) ([]models.URLResponse, error) {
	var result []models.URLResponse

	rows, err := db.Query(
		`SELECT short_url, original_url FROM urls 
         WHERE uuid = $1 AND is_deleted = false`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("database query error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var record models.URLResponse
		if err := rows.Scan(&record.ShortURL, &record.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, record)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return result, nil
}

func DeleteURLs(db *sql.DB, userID string) error {
	res, err := db.Exec(
		`UPDATE urls SET is_deleted = true WHERE user_id = $1 AND is_deleted = false`,
		userID,
	)

	fmt.Println(res)

	if err != nil {
		return err
	}

	return nil

}
