package control

import (
	"log"
	"net/http"
)

func (s *Server) handleBootstrapStart(w http.ResponseWriter, r *http.Request) {
	log.Printf("bootstrap start requested")
	status, err := s.deps.StartBootstrap(r.Context())
	if err != nil {
		log.Printf("bootstrap start failed: %v", err)
		writeError(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("bootstrap start accepted: state=%s step=%s", status.State, status.CurrentStep)
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}
