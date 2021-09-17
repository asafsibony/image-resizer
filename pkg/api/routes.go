package api

func (s *Server) initRoutes() {
	s.router.HandleFunc("/", s.handler)
	s.router.HandleFunc("/upload", s.uploadImage).Methods("POST")
}
