package api

func (s *Server) initRoutes() {
	s.router.HandleFunc("/", s.handler)
}
