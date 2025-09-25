package auth

import (
	"encoding/base64"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"testing"
	"time"
)

func TestCreateSignature(t *testing.T) {
	// Сохраняем оригинальный SecretKey
	originalSecretKey := config.Options.SecretKey
	defer func() {
		config.Options.SecretKey = originalSecretKey
	}()

	tests := []struct {
		name      string
		userID    string
		timestamp time.Time
		secretKey string
		wantError bool
	}{
		{
			name:      "Valid signature",
			userID:    "test-user-123",
			timestamp: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			secretKey: "test-secret-key",
			wantError: false,
		},
		{
			name:      "Empty user ID",
			userID:    "",
			timestamp: time.Now(),
			secretKey: "test-secret-key",
			wantError: false,
		},
		{
			name:      "Special characters in user ID",
			userID:    "user|123@example.com",
			timestamp: time.Now(),
			secretKey: "test-secret-key",
			wantError: false,
		},
		{
			name:      "Very long user ID",
			userID:    string(make([]byte, 1000)), // 1000 байт
			timestamp: time.Now(),
			secretKey: "test-secret-key",
			wantError: false,
		},
		{
			name:      "Empty secret key",
			userID:    "test-user",
			timestamp: time.Now(),
			secretKey: "",
			wantError: false,
		},
		{
			name:      "Zero timestamp",
			userID:    "test-user",
			timestamp: time.Time{},
			secretKey: "test-secret-key",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем тестовый секретный ключ
			config.Options.SecretKey = tt.secretKey

			// Вызываем тестируемую функцию
			signature := CreateSignature(tt.userID, tt.timestamp)

			// Проверяем что подпись не пустая
			if signature == "" {
				t.Error("Signature should not be empty")
			}

			// Проверяем что подпись является валидной base64 строкой
			_, err := base64.URLEncoding.DecodeString(signature)
			if err != nil {
				t.Errorf("Signature is not valid base64: %v", err)
			}

			// Проверяем детерминированность - одинаковые входные данные дают одинаковую подпись
			signature2 := CreateSignature(tt.userID, tt.timestamp)
			if signature != signature2 {
				t.Error("Signatures for same input should be identical")
			}

			// Проверяем что подпись зависит от всех входных параметров
			if tt.userID != "" {
				differentUserSig := CreateSignature(tt.userID+"-different", tt.timestamp)
				if signature == differentUserSig {
					t.Error("Signature should change when user ID changes")
				}
			}

			if !tt.timestamp.IsZero() {
				differentTimeSig := CreateSignature(tt.userID, tt.timestamp.Add(time.Second))
				if signature == differentTimeSig {
					t.Error("Signature should change when timestamp changes")
				}
			}

			if tt.secretKey != "" {
				// Временно меняем секретный ключ
				config.Options.SecretKey = tt.secretKey + "-different"
				differentKeySig := CreateSignature(tt.userID, tt.timestamp)
				config.Options.SecretKey = tt.secretKey // Восстанавливаем

				if signature == differentKeySig {
					t.Error("Signature should change when secret key changes")
				}
			}
		})
	}
}

func ExampleCreateSignature() {
	config.Options.SecretKey = "Some secret key"
	userID := "1234567"
	timestamp := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)

	signature := CreateSignature(userID, timestamp)
	fmt.Println(signature)

	// Output:
	// 8pVUemmqw2PB4garntayYEj69yjHWEzBpsmTT4cEGM0=
}

func ExampleCreateAuthCookie() {
	config.Options.SecretKey = "Some secret key"
	userID := "1234567"

	cookie, err := CreateAuthCookie(userID)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Auth name: " + cookie.Name)
		fmt.Printf("Maximum age: %d", cookie.MaxAge)
	}

	// Output:
	// Auth name: auth_user
	// Maximum age: 2592000

}
