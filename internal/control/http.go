package control

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func decodeBody(r *http.Request, target any) error {
	if r.Method != http.MethodPost {
		return fmt.Errorf("method %s is not allowed", r.Method)
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"error": map[string]string{
			"message": err.Error(),
		},
	})
}
