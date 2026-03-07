package control

import (
	"context"
	"net/http"
)

type SetupField struct {
	Key      string `json:"key"`
	Label    string `json:"label"`
	Required bool   `json:"required"`
	Status   string `json:"status"`
}

type SetupStatus struct {
	Complete       bool         `json:"complete"`
	BlockingFields []string     `json:"blockingFields,omitempty"`
	Required       []SetupField `json:"required"`
	Optional       []SetupField `json:"optional"`
}

func (s *Server) handleSetupFieldUpdate(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Key    string `json:"key"`
		Status string `json:"status"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	report, err := s.deps.UpdateSetupField(r.Context(), request.Key, request.Status)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": report})
}

type SetupFieldUpdater func(context.Context, string, string) (SetupStatus, error)
