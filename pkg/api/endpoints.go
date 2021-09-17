package api

import (
	"net/http"
)

func (s *Server) handler(w http.ResponseWriter, r *http.Request) {
	// vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("asaf"))
}
