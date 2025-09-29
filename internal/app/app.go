// Package app создает новый объект сервера с хендлером и роутером
package app

// Server объект сервера, имеющий хендлер и роутер
type Server struct {
	Handler *Handler
	Router  *Router
}

// NewServer Инициализирует сервер с пустым хендлером и роутером
func (s *Server) NewServer() *Server {
	serv := &Server{
		Handler: NewHandler(),
		Router:  NewRouter(),
	}

	return serv
}
