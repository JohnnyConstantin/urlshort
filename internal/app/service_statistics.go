package app

import (
	"database/sql"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"sync"
)

// DBStatistics объект статистики с использованием БД
type DBStatistics struct {
	db  *sql.DB
	cfg config.StorageConfig
}

func NewDBStatistics(db *sql.DB, cfg config.StorageConfig) *DBStatistics {
	return &DBStatistics{
		db:  db,
		cfg: cfg,
	}
}

// FileStatistics объект статистики с использованием файла
type FileStatistics struct {
	mu  *sync.Mutex
	cfg config.StorageConfig
}

func NewFileStatistics(cfg config.StorageConfig) *FileStatistics {
	return &FileStatistics{
		mu:  new(sync.Mutex),
		cfg: cfg,
	}
}

// MemoryStatistics объект статистики с использованием памяти
type MemoryStatistics struct {
	mu  *sync.Mutex
	cfg config.StorageConfig
}

func NewMemoryStatistics(cfg config.StorageConfig) *MemoryStatistics {
	return &MemoryStatistics{
		mu:  new(sync.Mutex),
		cfg: cfg,
	}
}

// GetUsersCount Получить количество пользователей для БД
func (s *DBStatistics) GetUsersCount() (int, error) {
	countUsers, err := store.GetUsersCount(s.db)
	if err != nil {
		return 0, err
	}
	return countUsers, nil
}

// GetUsersCount Получить количество пользователей для файла. По предыдущим спринтам пользователь в данном варианте отсутствует для обратной совместимости
func (s *FileStatistics) GetUsersCount() (int, error) {
	return 1, nil // 1 пользователь
}

// GetUsersCount Получить количество пользователей для памяти. По предыдущим спринтам пользователь в данном варианте отсутствует для обратной совместимости
func (s *MemoryStatistics) GetUsersCount() (int, error) {
	return 1, nil // 1 пользователь
}

// GetURLsCount Получить количество пURL для БД
func (s *DBStatistics) GetURLsCount() (int, error) {
	countURLS, err := store.GetURLsCount(s.db)
	if err != nil {
		return 0, err
	}
	return countURLS, nil
}

// GetURLsCount Получить количество URL для файла
func (s *FileStatistics) GetURLsCount() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(store.URLStore), nil
}

// GetURLsCount Получить количество URL для памяти
func (s *MemoryStatistics) GetURLsCount() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(store.URLStore), nil
}
