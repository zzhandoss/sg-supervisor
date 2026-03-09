package control

import "net/http"

func (s *Server) handleBootstrapStart(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.StartBootstrap(r.Context())
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}
