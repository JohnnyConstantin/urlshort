package app

type Server struct {
	Handler *Handler
	Router  *Router
}

func (s *Server) NewServer() *Server {
	serv := &Server{
		Handler: NewHandler(),
		Router:  NewRouter(),
	}

	return serv
}
