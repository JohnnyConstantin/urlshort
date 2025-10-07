// Package auth отвечает за создание сигнатур и кук для авторизации и аутентификации
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/JohnnyConstantin/urlshort/internal/config"
)

// CreateSignature создает подпись
func CreateSignature(userID string, timestamp time.Time) string {
	h := hmac.New(sha256.New, []byte(config.Options.SecretKey))
	data := fmt.Sprintf("%s|%d", userID, timestamp.Unix())
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// CreateAuthCookie создает куку
func CreateAuthCookie(userID string) (*http.Cookie, error) {
	now := time.Now()
	signature := CreateSignature(userID, now)

	value := fmt.Sprintf("%s|%d|%s", userID, now.Unix(), signature) // Задаем формат куки
	encoded := base64.URLEncoding.EncodeToString([]byte(value))

	return &http.Cookie{
		Name:     "auth_user",
		Value:    encoded,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
		Secure:   false, // Несколько часов пытался понять, почему кука не приходит - оказалось ждал HTTPS, вместо HTTP
		HttpOnly: true,  // Поэтому добавил это, чтобы наверняка
		SameSite: http.SameSiteLaxMode,
	}, nil
}
