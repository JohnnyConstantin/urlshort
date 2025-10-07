// Package config содержит основные параметры конфигурации,
// а также реализует логику выбора хранилища данных и парсинга переменных окружения
package config

import (
	"flag"
)

// StorageType тип хранилища
type StorageType string

// виды хранилища
const (
	StorageMemory StorageType = "memory"
	StorageFile   StorageType = "file"
	StorageDB     StorageType = "postgres"
)

// внутренние параметры для разработчика (предсказал их появление на первом же спринте xD)
//
//nolint:gochecknoglobals
var (
	AppName   = "shortener" // В дальнейшем может использоваться для CLI интерфейса
	PathToENV = ".env"      // Должен использоваться для подключения к БД
)

// Options опции запуска сервера
var Options struct {
	Address     string
	BaseAddress string
	DSN         string
	FileToWrite string
	SecretKey   string
}

// Config Объект глобального конфига
var Config StorageConfig

// StorageConfig Объект конфига хранилища
type StorageConfig struct {
	StorageType StorageType
	DatabaseDSN string // DSN для PostgreSQL (опциональное)
	FilePath    string // Путь к файлу (опциональное)
}

func init() {
	flag.StringVar(
		&Options.Address,
		"a",
		"localhost:8080",
		"The address to start the server on")
	flag.StringVar(
		&Options.BaseAddress,
		"b",
		"http://localhost:8080",
		"The address to return after shortener")
	flag.StringVar( // Странно, что локальные тесты намекают на необходимость этого флага в 6 инкременте, а реально появится он только в 9
		&Options.FileToWrite,
		"f",
		"",
		"File to write logs",
	)
	flag.StringVar( // DSN к БД
		&Options.DSN,
		"d",
		"",
		"Database connection string",
	)
	flag.StringVar( // Ключ для подписи куки
		&Options.SecretKey,
		"k",
		"default_key",
		"Secret key for user authentication",
	)
}

// CreateStorageConfig В зависимости от переданых параметров устанавливает StorageType для всего приложения
func CreateStorageConfig() {

	if Options.DSN != "" {
		Config = StorageConfig{
			StorageType: StorageDB,
			DatabaseDSN: Options.DSN,
		}
		return
	}

	// Fallback до StorageFile в случае, если СУБД не обнаружено
	if Options.FileToWrite != "" {
		Config = StorageConfig{
			StorageType: StorageFile,
			FilePath:    Options.FileToWrite,
		}
		return
	}

	// Fallback до inMemory в случае, если файл не обнаружен
	Config = StorageConfig{
		StorageType: StorageMemory,
	}
}

// GetStorageConfig геттер для конфига
func GetStorageConfig() StorageConfig {
	return Config
}
