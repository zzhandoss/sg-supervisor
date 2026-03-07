package control

import (
	"context"
	"net/http"
)

type ServiceHostArtifacts struct {
	ServiceName          string   `json:"serviceName"`
	DisplayName          string   `json:"displayName"`
	Description          string   `json:"description"`
	BinaryPath           string   `json:"binaryPath"`
	Arguments            []string `json:"arguments"`
	ListenAddress        string   `json:"listenAddress"`
	LinuxUnitPath        string   `json:"linuxUnitPath"`
	WindowsInstallPath   string   `json:"windowsInstallPath"`
	WindowsUninstallPath string   `json:"windowsUninstallPath"`
	WindowsStartPath     string   `json:"windowsStartPath"`
	WindowsStopPath      string   `json:"windowsStopPath"`
	WrittenFiles         []string `json:"writtenFiles"`
	InstallHints         []string `json:"installHints"`
	UninstallHints       []string `json:"uninstallHints"`
}

func (s *Server) handleServiceHostRender(w http.ResponseWriter, r *http.Request) {
	var request struct {
		BinaryPath string `json:"binaryPath"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	artifacts, err := s.deps.RenderServiceHostArtifacts(r.Context(), request.BinaryPath)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": artifacts})
}

type ServiceHostRenderer func(context.Context, string) (ServiceHostArtifacts, error)
