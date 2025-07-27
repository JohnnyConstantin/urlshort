package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
)

type DBDeleter struct {
	cfg config.StorageConfig
	db  *sql.DB
}

func (s *DBDeleter) DeleteURL(ctx context.Context, userID string, shortURLs []string) error {

	var batchSize = 100
	for i := 0; i < len(shortURLs); i += batchSize {
		end := i + batchSize
		if end > len(shortURLs) {
			end = len(shortURLs)
		}

		batch := shortURLs[i:end] // Задел под удаление множества записей (1000+), чтобы не повесить бд

		fmt.Println(batch)
		err := store.DeleteURLs(s.db, userID, batch)
		if err != nil {
			return err
		}
	}
	return nil

}
