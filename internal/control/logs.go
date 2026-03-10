package control

import (
	"net/http"
	"strconv"
)

type RecentLogsResponse struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

func (s *Server) handleRecentLogs(w http.ResponseWriter, r *http.Request) {
	limit := 80
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	logs, err := s.deps.ReadRecentLogs(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": logs})
}
