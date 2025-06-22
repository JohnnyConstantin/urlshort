package app

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
