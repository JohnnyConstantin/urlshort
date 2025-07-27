package app

import (
	"context"
	"database/sql"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
	"sync"
)

type DBDeleter struct {
	cfg config.StorageConfig
	db  *sql.DB
}

func (s *DBDeleter) DeleteURL(ctx context.Context, userID string, shortURLs []string) error {

	if len(shortURLs) == 0 {
		return nil
	}

	// Канал для входящих URL
	inputChan := make(chan string, len(shortURLs))

	// Заполняем канал
	go func() {
		defer close(inputChan)
		for _, url := range shortURLs {
			inputChan <- url
		}
	}()

	// Канал для ошибок
	errChan := make(chan error, 1)

	// Запускаем worker'ов (оптимальное количество - обычно 2-4x CPU cores)
	const workerCount = 4
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Канал для батчей
	batchChan := make(chan []string, workerCount)

	// Fan-out: распределяем работу по worker'ам
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			worker(ctx, s.db, userID, inputChan, batchChan, errChan)
		}()
	}

	// Fan-in: собираем результаты
	go func() {
		wg.Wait()
		close(batchChan)
		close(errChan)
	}()

	// Обрабатываем результаты
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil

}

func worker(ctx context.Context, db *sql.DB, userID string,
	inputChan <-chan string, batchChan chan<- []string, errChan chan<- error) {

	const batchSize = 100
	batch := make([]string, 0, batchSize)

	for url := range inputChan {
		batch = append(batch, url)

		if len(batch) >= batchSize {
			if err := store.DeleteURLs(db, userID, batch); err != nil {
				errChan <- err
				return
			}
			batch = batch[:0] // Сбрасываем батч
		}
	}

	// Обрабатываем оставшиеся элементы
	if len(batch) > 0 {
		if err := store.DeleteURLs(db, userID, batch); err != nil {
			errChan <- err
			return
		}
	}
}
