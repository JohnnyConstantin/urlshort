package store

import (
	"database/sql"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/models"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func GetDB() (*sql.DB, error) {

	//Вгружаем переменные окружения, в т.ч. креды для бд
	err := godotenv.Load(config.PATH_TO_ENV)
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	pgDB := os.Getenv("PG_DATABASE")
	pgUser := os.Getenv("PG_USER")
	pgPassword := os.Getenv("PG_PASSWORD")
	pgHost := os.Getenv("PG_HOST")
	pgPort := os.Getenv("PG_PORT")

	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=disable",
		pgUser, pgDB, pgPassword, pgHost, pgPort)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func IsDuplicate(db *sql.DB, originalURL string) (string, bool) {
	var existingShortID string

	err := db.QueryRow(
		`SELECT short_id FROM urls WHERE original_url = $1 LIMIT 1`,
		originalURL,
	).Scan(&existingShortID)

	if err == nil {
		return existingShortID, true
	} else {
		return "", false
	}
}

func Insert(db *sql.DB, originalURL string) (string, error) {
	var result string
	shortID := uuid.New().String()[:8]

	_, err := db.Exec(
		"INSERT INTO urls (short_id, original_url) VALUES ($1, $2)",
		shortID,
		originalURL,
	)
	if err != nil {
		return "", err
	}

	result = shortID

	return result, nil
}

func Read(db *sql.DB, shortID string) (string, error) {
	var model models.URLResponse

	row := db.QueryRow(
		`SELECT short_id, original_url 
		FROM urls 
		WHERE short_id = $1`,
		shortID,
	)

	err := row.Scan(&model.ShortURL, &model.OriginalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return model.OriginalURL, err
		}
		return model.OriginalURL, err
	}

	return model.OriginalURL, nil
}
