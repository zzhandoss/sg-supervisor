package control

import (
	"context"
	"net/http"
)

type ProductConfigStatus struct {
	PreferredHost                   string   `json:"preferredHost,omitempty"`
	ResolvedHost                    string   `json:"resolvedHost"`
	AvailableHosts                  []string `json:"availableHosts"`
	TelegramBotConfigured           bool     `json:"telegramBotConfigured"`
	ViteAPIBaseURL                  string   `json:"viteApiBaseUrl"`
	AdminUIURL                      string   `json:"adminUiUrl"`
	APICorsAllowedOrigins           []string `json:"apiCorsAllowedOrigins"`
	DeviceServiceCorsAllowedOrigins []string `json:"deviceServiceCorsAllowedOrigins"`
}

type ProductConfigUpdate struct {
	PreferredHost         *string `json:"preferredHost,omitempty"`
	TelegramBotToken      *string `json:"telegramBotToken,omitempty"`
	ClearTelegramBotToken bool    `json:"clearTelegramBotToken,omitempty"`
}

func (s *Server) handleProductConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var request ProductConfigUpdate
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	status, err := s.deps.UpdateProductConfig(r.Context(), request)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

type ProductConfigUpdater func(context.Context, ProductConfigUpdate) (ProductConfigStatus, error)
