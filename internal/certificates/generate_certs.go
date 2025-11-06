package certificates

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"time"
)

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
