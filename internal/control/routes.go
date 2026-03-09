package control

import "net/http"

func (s *Server) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/api/v1/activation-request", s.handleActivationRequest)
	mux.HandleFunc("/api/v1/license/import", s.handleLicenseImport)
	mux.HandleFunc("/api/v1/services/start", s.handleServiceStart)
	mux.HandleFunc("/api/v1/services/stop", s.handleServiceStop)
	mux.HandleFunc("/api/v1/services/restart", s.handleServiceRestart)
	mux.HandleFunc("/api/v1/updates/import-manifest", s.handleManifestImport)
	mux.HandleFunc("/api/v1/updates/import-bundle", s.handleBundleImport)
	mux.HandleFunc("/api/v1/updates/apply-local-bundle", s.handleApplyLocalBundle)
	mux.HandleFunc("/api/v1/updates/apply", s.handleApplyPackage)
	mux.HandleFunc("/api/v1/setup/fields", s.handleSetupFieldUpdate)
	mux.HandleFunc("/api/v1/product-config", s.handleProductConfigUpdate)
	mux.HandleFunc("/api/v1/install", s.handleInstallPackage)
	mux.HandleFunc("/api/v1/repair", s.handleRepair)
	mux.HandleFunc("/api/v1/uninstall", s.handleUninstall)
	mux.HandleFunc("/api/v1/service-host/render", s.handleServiceHostRender)
	mux.HandleFunc("/api/v1/manifests/validate", s.handleManifestValidation)
	mux.HandleFunc("/", handleUI)
	return mux
}
