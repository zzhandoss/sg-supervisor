package control

import "net/http"

func writeErrorWithData(w http.ResponseWriter, status int, err error, data any) {
	writeJSON(w, status, map[string]any{
		"success": false,
		"data":    data,
		"error": map[string]string{
			"message": err.Error(),
		},
	})
}
