package app

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	auth "github.com/JohnnyConstantin/urlshort/auth"
)

type myKeyType string

const (
	user myKeyType = "user"
)

// WithAuth мидлварь, которая осуществляет аутентификацию к последующему хендлеру
func (h *Handler) WithAuth(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Попытка аутентификации по куке
		cookie, err := r.Cookie("auth_user")
		if err == nil { // если кука есть
			decoded, err := base64.URLEncoding.DecodeString(cookie.Value) // вытаскиваем и декодим
			if err != nil {                                               // если ошибка в кодировке - выкидываем ошибку
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			parts := strings.Split(string(decoded), "|") // Если нет трех |, значит неверный формат куки
			if len(parts) != 3 {
				http.Error(w, errors.New("invalid cookie").Error(), http.StatusBadRequest)
				return
			}

			// Вытаскиываем части куки для верификации
			userID := parts[0]
			timestampStr := parts[1]
			signature := parts[2]

			if userID == "" { // Кажется, что затриггерить это невозможно, потому что в тестах клиент получает куку от сервера
				http.Error(w, "No such user", http.StatusUnauthorized) // 401 Unauthorized
				return
			}

			timestampInt := int64(0)
			_, err = fmt.Sscanf(timestampStr, "%d", &timestampInt)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			timestamp := time.Unix(timestampInt, 0)

			expectedSig := auth.CreateSignature(userID, timestamp)
			if !hmac.Equal([]byte(signature), []byte(expectedSig)) { // Если неверная подпись, выдаем новую куку
				userID, err = authenticate(w)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			ctx := context.WithValue(r.Context(), user, userID)
			// Прокидываем дальше
			hf(w, r.WithContext(ctx))

		} else { // если куки нет, то авторизовать
			newUserID, err := authenticate(w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := context.WithValue(r.Context(), user, newUserID)
			// Прокидываем дальше
			hf(w, r.WithContext(ctx))
		}
	}
}

func authenticate(w http.ResponseWriter) (string, error) {
	userID := uuid.New().String()
	newCookie, err := auth.CreateAuthCookie(userID)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, newCookie)
	return userID, nil
}
