package app

import (
	"database/sql"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
)

type DBDeleter struct {
	cfg config.StorageConfig
	db  *sql.DB
}

func (s *DBDeleter) DeleteURL(userID string) error {
	err := store.DeleteURLs(s.db, userID)
	if err != nil {
		return err
	}

	return nil
}
