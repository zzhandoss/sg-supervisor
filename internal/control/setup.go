package control

import (
	"context"
	"log"
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
		Value  string `json:"value,omitempty"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("setup field update requested: key=%s status=%s", request.Key, request.Status)
	report, err := s.deps.UpdateSetupField(r.Context(), request.Key, request.Status, request.Value)
	if err != nil {
		log.Printf("setup field update failed for %s: %v", request.Key, err)
		writeError(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("setup field updated: key=%s status=%s", request.Key, request.Status)
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": report})
}

type SetupFieldUpdater func(context.Context, string, string, string) (SetupStatus, error)
