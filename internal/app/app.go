// Package app создает новый объект сервера с хендлером и роутером
package app

import (
	"context"
	"database/sql"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server объект сервера, имеющий хендлер и роутер
type Server struct {
	Handler    *Handler
	Router     *Router
	HTTPServer *http.Server
	DB         *sql.DB
}

// NewServer Инициализирует сервер с пустым хендлером и роутером
func (s *Server) NewServer(service *Service) *Server {
	serv := &Server{
		Handler: NewHandler(service),
		Router:  NewRouter(),
	}

	return serv
}

// Start запускает HTTP сервер и возвращает канал для ожидания завершения
func (s *Server) Start(addr string, router *chi.Mux) error {
	s.HTTPServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return s.HTTPServer.ListenAndServe()
}

// StartTLS запускает HTTPS сервер и возвращает канал для ожидания завершения
func (s *Server) StartTLS(addr, certFile, keyFile string, router *chi.Mux) error {
	s.HTTPServer = &http.Server{
		Addr:    addr,
		Handler: router,
	}

	return s.HTTPServer.ListenAndServeTLS(certFile, keyFile)
}

// WaitForShutdown ожидает сигналов завершения работы
func (s *Server) WaitForShutdown(logger zap.SugaredLogger) error {
	// Канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	// Блокируемся до получения сигнала
	<-sigChan

	logger.Info("Received shutdown signal...")

	// Инициируем graceful shutdown
	err := s.Shutdown()
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Shutdown() error {

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Останавливаем HTTP сервер
	if err := s.HTTPServer.Shutdown(ctx); err != nil {
		return err
	}

	// Закрываем соединение с базой данных. Close самостоятельно дожидается окончания всех начатых операций с БД
	if s.DB != nil {
		if err := s.DB.Close(); err != nil {
			return err
		}
	}

	return nil
}
