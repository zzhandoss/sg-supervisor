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
	WindowsWrapperPath   string   `json:"windowsWrapperPath,omitempty"`
	WindowsConfigPath    string   `json:"windowsConfigPath,omitempty"`
	LinuxUnitPath        string   `json:"linuxUnitPath"`
	WindowsInstallPath   string   `json:"windowsInstallPath"`
	WindowsUninstallPath string   `json:"windowsUninstallPath"`
	WindowsStartPath     string   `json:"windowsStartPath"`
	WindowsStopPath      string   `json:"windowsStopPath"`
	WrittenFiles         []string `json:"writtenFiles"`
	InstallHints         []string `json:"installHints"`
	UninstallHints       []string `json:"uninstallHints"`
}

type ServiceHostStatus struct {
	Supported   bool   `json:"supported"`
	ServiceName string `json:"serviceName"`
	Installed   bool   `json:"installed"`
	State       string `json:"state"`
	StartMode   string `json:"startMode"`
	WrapperPath string `json:"wrapperPath,omitempty"`
	ConfigPath  string `json:"configPath,omitempty"`
	LastError   string `json:"lastError,omitempty"`
	Description string `json:"description,omitempty"`
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

func (s *Server) handleServiceHostStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.ServiceHostStatus(r.Context())
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostInstall(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.InstallServiceHost(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostStart(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.StartServiceHost(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostSwitch(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.SwitchToServiceHost(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostStop(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.StopServiceHost(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostEnableAutostart(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.EnableServiceHostAutostart(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostDisableAutostart(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.DisableServiceHostAutostart(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

func (s *Server) handleServiceHostRemove(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.RemoveServiceHost(r.Context())
	if err != nil {
		writeErrorWithData(w, http.StatusBadRequest, err, status)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": status})
}

type ServiceHostRenderer func(context.Context, string) (ServiceHostArtifacts, error)
type ServiceHostReader func(context.Context) (ServiceHostStatus, error)
type ServiceHostMutator func(context.Context) (ServiceHostStatus, error)
