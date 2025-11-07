package grpcserver

import (
	"context"
	"database/sql"
	"errors"
	service "github.com/JohnnyConstantin/urlshort/internal/app"
	"github.com/JohnnyConstantin/urlshort/internal/config"
	"github.com/JohnnyConstantin/urlshort/internal/store"
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
	var err error

	id := req.GetShortUrl()

	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		s.service.Fuller = &service.FileFuller{Cfg: cfg}
		s.service.Fuller.InitMutex()
		response, exists, isDeleted, err = s.service.Fuller.GetFullURL(id)
		if err != nil {
			return nil, err
		}
	case config.StorageMemory:
		s.service.Fuller = &service.MemoryFuller{Cfg: cfg}
		s.service.Fuller.InitMutex()
		response, exists, isDeleted, err = s.service.Fuller.GetFullURL(id)
		if err != nil {
			return nil, err
		}
	case config.StorageDB:
		s.service.Fuller = &service.DBFuller{Db: s.Db, Cfg: cfg}
		response, exists, isDeleted, err = s.service.Fuller.GetFullURL(id)
		if err != nil {
			return nil, err
		}
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
		ShortUrl:    req.GetShortUrl(),
	}, nil
}

func (s *GRPCServer) CreateShortURLBatch(ctx context.Context, req *shortener.CreateShortURLBatchRequest) (*shortener.CreateShortURLBatchResponse, error) {

	requests := req.GetUrls()

	cfg := config.GetStorageConfig()

	responses := make([]*shortener.URLPair, 0, len(requests))

	switch cfg.StorageType {
	case config.StorageFile:
		s.service.Shortener = &service.FileShortener{Cfg: cfg}
		s.service.Shortener.InitMutex()

	case config.StorageMemory:
		s.service.Shortener = &service.MemoryShortener{Cfg: cfg}
		s.service.Shortener.InitMutex()

	case config.StorageDB:
		s.service.Shortener = &service.DBShortener{Db: s.Db, Cfg: cfg}
	default:
		return nil, errors.New("invalid storage type")
	}

	for _, r := range requests {
		shorten_req := service.Shortenerequest{OriginalURL: r.OriginalUrl}
		ShortURL := s.service.Shortener.ShortenURL(shorten_req)

		responses = append(responses, &shortener.URLPair{
			CorrelationId: r.GetCorrelationId(),
			ShortUrl:      ShortURL.Result,
			OriginalUrl:   "", // Возвращаем пустое значение
		})
	}

	return &shortener.CreateShortURLBatchResponse{
		Urls: responses,
	}, nil
}

func (s *GRPCServer) GetOriginalURLBatch(ctx context.Context, req *shortener.GetOriginalURLBatchRequest) (*shortener.GetOriginalURLBatchResponse, error) {

	userID := req.GetUserId()

	cfg := config.GetStorageConfig()

	var response []*shortener.URLPair

	switch cfg.StorageType {
	case config.StorageFile:
		s.service.Fuller = &service.FileFuller{Cfg: cfg}
		s.service.Fuller.InitMutex()

	case config.StorageMemory:
		s.service.Fuller = &service.MemoryFuller{Cfg: cfg}
		s.service.Fuller.InitMutex()

	case config.StorageDB:
		s.service.Fuller = &service.DBFuller{Db: s.Db, Cfg: cfg}
	default:
		return nil, errors.New("invalid storage type")
	}

	urls, err := store.ReadWithUUID(s.Db, userID)
	if err != nil {
		return nil, errors.New("error while reading user urls")
	}

	if len(urls) == 0 {
		return nil, errors.New("no content")
	}

	for _, url := range urls {
		response = append(response, &shortener.URLPair{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		})
	}

	return &shortener.GetOriginalURLBatchResponse{
		Urls: response,
	}, nil
}

func (s *GRPCServer) DeleteUserURLs(ctx context.Context, req *shortener.DeleteUserURLsRequest) (*shortener.DeleteUserURLsResponse, error) {
	userID := req.GetUserId()
	shortURLs := req.GetUrls()
	cfg := config.GetStorageConfig()

	deleter := service.DBDeleter{Cfg: cfg, Db: s.Db}

	err := deleter.DeleteURL(userID, shortURLs)
	if err != nil {
		return nil, err
	}

	return &shortener.DeleteUserURLsResponse{
		Result: "Successfully deleted",
	}, nil

}

func (s *GRPCServer) GetStats(ctx context.Context, req *shortener.GetStatsRequest) (*shortener.GetStatsResponse, error) {
	var statistics service.Statter

	cfg := config.GetStorageConfig()

	switch cfg.StorageType {
	case config.StorageFile:
		statistics = service.NewFileStatistics(cfg)

	case config.StorageMemory:
		statistics = service.NewMemoryStatistics(cfg)

	case config.StorageDB:
		statistics = service.NewDBStatistics(s.Db, cfg)

	default: // Overkill, но перестраховаться нужно
		return nil, errors.New("invalid storage type")
	}

	cnt, err := statistics.GetURLsCount()
	if err != nil {
		return nil, err
	}
	usrs, err := statistics.GetUsersCount()
	if err != nil {
		return nil, err
	}

	return &shortener.GetStatsResponse{
		UrlsCount:  int64(usrs),
		UsersCount: int64(cnt),
	}, nil
}

func (s *GRPCServer) Ping(ctx context.Context, req *shortener.PingRequest) (*shortener.PingResponse, error) {
	var result bool
	err := s.Db.Ping()
	if err != nil {
		result = false
		return &shortener.PingResponse{Success: result}, err
	}

	result = true
	return &shortener.PingResponse{Success: result}, nil
}
