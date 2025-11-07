package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	service "github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/models"
	shortener "github.com/JohnnyConstantin/urlshort/shortener/proto"
)

type GRPCServer struct {
	shortener.UnimplementedShortenerServer
	service *service.Service
	Db      *sql.DB
}

func NewGRPCServer(service *service.Service) *GRPCServer {
	return &GRPCServer{
		service: service,
	}
}

// В архитектуре заранее был заложен абстрактный слой сервиса, поэтому просто повторяем обращения к
// сервисам с небольшими правками, как и в HTTP хендлерах
func (s *GRPCServer) CreateShortURL(ctx context.Context, r *shortener.CreateShortURLRequest) (*shortener.CreateShortURLResponse, error) {

	var ShortURL models.ShortenResponse

	cfg := config.GetStorageConfig()
	switch cfg.StorageType {
	case config.StorageFile:
		s.service.Shortener = &service.FileShortener{Cfg: cfg}
		s.service.Shortener.InitMutex()
		shorten_req := service.Shortenerequest{OriginalURL: r.OriginalUrl}
		ShortURL = s.service.Shortener.ShortenURL(shorten_req)
	case config.StorageMemory:
		s.service.Shortener = &service.MemoryShortener{Cfg: cfg}
		s.service.Shortener.InitMutex()
		shorten_req := service.Shortenerequest{OriginalURL: r.OriginalUrl}
		ShortURL = s.service.Shortener.ShortenURL(shorten_req)
	case config.StorageDB:
		s.service.Shortener = &service.DBShortener{Db: s.Db, Cfg: cfg}
		short := service.DBShortener{Db: s.Db, Cfg: cfg}
		shorten_req := service.Shortenerequest{OriginalURL: r.OriginalUrl, UserID: r.UserId}

		ShortURL = short.ShortenURL(shorten_req)
	default:
		return nil, errors.New("invalid storage type")
	}

	return &shortener.CreateShortURLResponse{ShortUrl: ShortURL.Result}, nil
}

func (s *GRPCServer) GetOriginalURL(ctx context.Context, req *shortener.GetOriginalURLRequest) (*shortener.GetOriginalURLResponse, error) {

	response := models.ShortenRequest{URL: ""}
	exists := false //By default не существует
	var isDeleted bool

	id := req.GetShortUrlId()

	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		s.service.Fuller = &service.FileFuller{Cfg: cfg}
		s.service.Fuller.InitMutex()
		response, exists, _ = s.service.Fuller.GetFullURL(id)
	case config.StorageMemory:
		s.service.Fuller = &service.MemoryFuller{Cfg: cfg}
		s.service.Fuller.InitMutex()
		response, exists, _ = s.service.Fuller.GetFullURL(id)
	case config.StorageDB:
		s.service.Fuller = &service.DBFuller{Db: s.Db, Cfg: cfg}
		response, exists, isDeleted = s.service.Fuller.GetFullURL(id)
		if isDeleted {
			return nil, errors.New("this URL was deleted")
		}
	default:
		return nil, errors.New("invalid storage type")
	}

	if !exists {
		return nil, errors.New("this URL does not exist")
	}

	return &shortener.GetOriginalURLResponse{
		OriginalUrl: response.URL,
		ShortUrl:    req.ShortUrlId,
	}, nil
}
