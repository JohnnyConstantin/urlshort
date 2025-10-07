package app

import (
	"database/sql"
	"sync"

	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
)

// Возможно в будущем появятся разные реализации удаления
type DBDeleter struct {
	cfg config.StorageConfig
	db  *sql.DB
}

// DeleteURL удалить URL в БД
func (s *DBDeleter) DeleteURL(userID string, shortURLs []string) error {

	if len(shortURLs) == 0 {
		return nil
	}

	// Канал для входящих URL
	inputChan := make(chan string, len(shortURLs))

	// Заполнение канала
	go func() {
		defer close(inputChan)
		for _, url := range shortURLs {
			inputChan <- url
		}
	}()

	// Канал для ошибок
	errChan := make(chan error, 1)

	// Запускаем worker (4 шт выбрал наугад)
	const workerCount = 4
	var wg sync.WaitGroup
	wg.Add(workerCount)

	// Fan-out. Раскидываем горутины по воркерам
	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			worker(s.db, userID, inputChan, errChan)
		}()
	}

	// Fan-in. собираем результаты. Inputchan закрывается ранее
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Обработка ошибок
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil

}

func worker(db *sql.DB, userID string,
	inputChan <-chan string, errChan chan<- error) {

	const batchSize = 100 //наверное многовато, но если сделать меньше, то смысла в батчах как-будто вообще не будет
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
