package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"net/http"
	"time"
)

func CreateSignature(userID string, timestamp time.Time) string {
	h := hmac.New(sha256.New, []byte(config.Options.SecretKey))
	data := fmt.Sprintf("%s|%d", userID, timestamp.Unix())
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func CreateAuthCookie(userID string) (*http.Cookie, error) {
	now := time.Now()
	signature := CreateSignature(userID, now)

	value := fmt.Sprintf("%s|%d|%s", userID, now.Unix(), signature)
	encoded := base64.URLEncoding.EncodeToString([]byte(value))

	return &http.Cookie{
		Name:     "auth_user",
		Value:    encoded,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}
