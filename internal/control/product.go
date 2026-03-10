package control

import (
	"context"
	"log"
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
	log.Printf("product config update requested")
	status, err := s.deps.UpdateProductConfig(r.Context(), request)
	if err != nil {
		log.Printf("product config update failed: %v", err)
		writeError(w, http.StatusBadRequest, err)
		return
	}
	log.Printf("product config updated: preferredHost=%s telegramConfigured=%t", status.PreferredHost, status.TelegramBotConfigured)
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

type ProductConfigUpdater func(context.Context, ProductConfigUpdate) (ProductConfigStatus, error)
