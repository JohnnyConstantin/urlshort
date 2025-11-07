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
}

func NewGRPCServer(service *service.Service) *GRPCServer {
	return &GRPCServer{
		service: service,
	}
}

// В архитектуре заранее был заложен абстрактный слой сервиса, поэтому буквально повторяем обращения к сервисам, как и в HTTP хендлерах
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
		db, userID, errs := initCtx(ctx)
		if errs != nil {
			return nil, errs
		}

		s.service.Shortener = &service.DBShortener{Db: db, Cfg: cfg}
		short := service.DBShortener{Db: db, Cfg: cfg}
		shorten_req := service.Shortenerequest{OriginalURL: r.OriginalUrl, UserID: userID}

		ShortURL = short.ShortenURL(shorten_req)
	default:
		return nil, errors.New("invalid storage type")
	}

	return &shortener.CreateShortURLResponse{ShortUrl: ShortURL.Result}, nil
}

//func (s *GRPCServer) GetOriginalURL(ctx context.Context, req *shortener.GetOriginalURLRequest) (*shortener.GetOriginalURLResponse, error) {
//	originalURL, err := s.service.GetOriginalURL(ctx, req.ShortUrlId)
//	if err != nil {
//		return nil, err
//	}
//
//	return &shortener.GetOriginalURLResponse{
//		OriginalUrl: originalURL,
//		ShortUrl:    s.service.BuildShortURL(req.ShortUrlId),
//	}, nil
//}
//
//// Аналогично реализуйте остальные методы...
//
//func (s *GRPCServer) getUserIDFromContext(ctx context.Context) string {
//	// Логика извлечения userID из контекста (аналог WithAuth)
//	return "user-id-from-token"
//}

// Здесь своя приватная инициализация контекста, поскольку с HTTP они отличаются
func initCtx(ctx context.Context) (*sql.DB, string, error) {
	db, ok := ctx.Value(service.DbKey).(*sql.DB)
	if !ok {
		return nil, "", errors.New("DB not in context")
	}
	userID, ok := ctx.Value(service.User).(string)
	if !ok {
		return nil, "", errors.New("userID not found in context")
	}

	return db, userID, nil
}
