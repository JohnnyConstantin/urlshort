package app

import (
	"github.com/JohnnyConstantin/urlshort/models"
)

type Shortenerequest struct {
	OriginalURL string
	UserID      string
}

// Shortener интерфейс для разных функций бизнес-логики в зависимости от используемого StorageType
type Shortener interface {
	ShortenURL(opts Shortenerequest) models.ShortenResponse // Используется для сжатия URL, используя оригинал
	InitMutex()
}

type Fuller interface {
	GetFullURL(id string) (models.ShortenRequest, bool, bool, error)
	InitMutex()
}

// Общая "точка входа" (витрина) для shortener и fuller
type Service struct {
	Shortener Shortener
	Fuller    Fuller
}

// Statter интерфейс для разных функций cтатистики в зависимости от используемого StorageType
type Statter interface {
	GetUsersCount() (int, error) // Получить количество пользователей
	GetURLsCount() (int, error)  // Получить количество URL
}
