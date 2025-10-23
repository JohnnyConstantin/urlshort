// Package config содержит основные параметры конфигурации,
// а также реализует логику выбора хранилища данных и парсинга переменных окружения
package config

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"
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
	EnableHTTPS bool // Добавлена опция на HTTPS
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
	flag.BoolVar( // Ключ для HTTPS
		&Options.EnableHTTPS,
		"s",
		true,
		"Enable HTTPS server",
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

// GenerateCertAndPrivFiles Генерация ключа и сертификата
func GenerateCertAndPrivFiles(certFile, keyFile string) error {
	// Генерация приватного ключа
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %v", err)
	}

	// Создание шаблона сертификата для генерации, если его нет в директории
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		// Информация о субъекте
		Subject: pkix.Name{
			Organization:  []string{"GOogle"},
			Country:       []string{"RU"},
			Province:      []string{""},
			Locality:      []string{""},
			StreetAddress: []string{""},
			PostalCode:    []string{""},
			CommonName:    "localhost",
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(365 * 24 * time.Hour),
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", certFile, err)
	}

	defer func(certOut *os.File) {
		err = certOut.Close()
		if err != nil {
			return
		}
	}(certOut)

	if err = pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", certFile, err)
	}

	keyOut, err := os.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", keyFile, err)
	}
	defer func(keyOut *os.File) {
		err = keyOut.Close()
		if err != nil {
			return
		}
	}(keyOut)

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %v", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privBytes,
	}); err != nil {
		return fmt.Errorf("failed to write data to %s: %v", keyFile, err)
	}

	return nil
}

// СertFilesExist проверяет существование файлов сертификата и ключа
func СertFilesExist(certFile, keyFile string) bool {
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(keyFile); os.IsNotExist(err) {
		return false
	}
	return true
}
