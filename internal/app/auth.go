package app

import (
	"context"
	"crypto/hmac"
	"encoding/base64"
	"errors"
	"fmt"
	auth "github.com/JohnnyConstantin/urlshort/internal/auth"
	"github.com/google/uuid"
	"net/http"
	"strings"
	"time"
)

type myKeyType string

const (
	user myKeyType = "user"
)

func (h *Handler) WithAuth(hf http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Попытка аутентификации по куке
		cookie, err := r.Cookie("auth_user")
		if err == nil { // если кука есть
			decoded, err := base64.URLEncoding.DecodeString(cookie.Value) // проверка кодировки
			if err != nil {                                               // если кодировка неверная - выдача новой куки
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			parts := strings.Split(string(decoded), "|")
			if len(parts) != 3 {
				http.Error(w, errors.New("invalid cookie").Error(), http.StatusBadRequest)
				return
			}

			// Вытаскиываем части куки для верификации
			userID := parts[0]
			timestampStr := parts[1]
			signature := parts[2]

			fmt.Println("USRID: " + userID)
			if userID == "" {
				http.Error(w, "No such user", http.StatusUnauthorized)
				return
			}

			timestampInt := int64(0)
			fmt.Sscanf(timestampStr, "%d", &timestampInt)
			timestamp := time.Unix(timestampInt, 0)

			expectedSig := auth.CreateSignature(userID, timestamp)
			if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
				userID = authenticate(w)
			}

			println("Using cookie:" + userID)
			ctx := context.WithValue(r.Context(), user, userID)
			// Прокидываем дальше
			hf(w, r.WithContext(ctx))

		} else { // если куки нет, то авторизовать
			newUserID := authenticate(w)
			ctx := context.WithValue(r.Context(), user, newUserID)
			fmt.Println("Using cookie:" + newUserID)
			// Прокидываем дальше
			hf(w, r.WithContext(ctx))
		}
	}
}

func authenticate(w http.ResponseWriter) string {
	userID := uuid.New().String()
	newCookie, err := auth.CreateAuthCookie(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.SetCookie(w, newCookie)
	return userID
}
