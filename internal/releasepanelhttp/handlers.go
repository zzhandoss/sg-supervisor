package releasepanelhttp

import (
	"net/http"

	"sg-supervisor/internal/releasepanel"
)

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]string{"status": "ok"}})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.service.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": status})
}

func (s *Server) handleRecipeUpdate(w http.ResponseWriter, r *http.Request) {
	var request releasepanel.Recipe
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	status, err := s.service.UpdateRecipe(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": status})
}

func (s *Server) handleVersions(w http.ResponseWriter, r *http.Request) {
	versions, err := s.service.ListVersions(r.Context(), r.URL.Query().Get("repo"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": versions})
}

func (s *Server) handleLocalRelease(w http.ResponseWriter, r *http.Request) {
	job, err := s.service.StartLocalRelease(r.Context())
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": job})
}

func (s *Server) handleIssueLicense(w http.ResponseWriter, r *http.Request) {
	var request releasepanel.LicenseIssueRequest
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	record, err := s.service.IssueLicense(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": record})
}
