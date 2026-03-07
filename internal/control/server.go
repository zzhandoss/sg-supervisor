package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sg-supervisor/internal/license"
	"sg-supervisor/internal/runtime"
)

type DirectoryStatus struct {
	Install  string `json:"install"`
	Config   string `json:"config"`
	Data     string `json:"data"`
	Logs     string `json:"logs"`
	Licenses string `json:"licenses"`
	Backups  string `json:"backups"`
	Runtime  string `json:"runtime"`
	Updates  string `json:"updates"`
}

type StatusResponse struct {
	ProductName      string                  `json:"productName"`
	Root             string                  `json:"root"`
	ListenAddr       string                  `json:"listenAddr"`
	Directories      DirectoryStatus         `json:"directories"`
	SetupRequired    bool                    `json:"setupRequired"`
	Setup            SetupStatus             `json:"setup"`
	License          license.Status          `json:"license"`
	LastUpdate       UpdateOperationStatus   `json:"lastUpdate"`
	ManagedServices  []string                `json:"managedServices"`
	Services         []runtime.ServiceStatus `json:"services"`
	ImportedPackages []PackageRecord         `json:"importedPackages"`
	ActivePackage    ActivePackageRecord     `json:"activePackage"`
}

type PackageRecord struct {
	PackageID         string   `json:"packageId"`
	SourceType        string   `json:"sourceType"`
	SourcePath        string   `json:"sourcePath"`
	ArchivePath       string   `json:"archivePath,omitempty"`
	StoredPath        string   `json:"storedPath"`
	StageDir          string   `json:"stageDir,omitempty"`
	ImportedAt        string   `json:"importedAt"`
	ProductVersion    string   `json:"productVersion"`
	CoreVersion       string   `json:"coreVersion"`
	SupervisorVersion string   `json:"supervisorVersion"`
	Adapters          []string `json:"adapters"`
}

type ActivePackageRecord struct {
	PackageID         string   `json:"packageId,omitempty"`
	AppliedAt         string   `json:"appliedAt,omitempty"`
	ProductVersion    string   `json:"productVersion,omitempty"`
	CoreVersion       string   `json:"coreVersion,omitempty"`
	SupervisorVersion string   `json:"supervisorVersion,omitempty"`
	ManifestPath      string   `json:"manifestPath,omitempty"`
	BackupPath        string   `json:"backupPath,omitempty"`
	Adapters          []string `json:"adapters,omitempty"`
	StoppedServices   []string `json:"stoppedServices,omitempty"`
	RestartedServices []string `json:"restartedServices,omitempty"`
}

func (s StatusResponse) PrettyString() string {
	return fmt.Sprintf(
		"product=%s setupRequired=%t licenseValid=%t root=%s listen=%s",
		s.ProductName,
		s.SetupRequired,
		s.License.Valid,
		s.Root,
		s.ListenAddr,
	)
}

type HandlerDependencies struct {
	Status                     func(context.Context) (StatusResponse, error)
	GenerateActivationRequest  func(context.Context, string, string) (string, error)
	ImportLicense              func(context.Context, string) error
	StartService               func(context.Context, string) error
	StopService                func(string) error
	RestartService             func(context.Context, string) error
	ImportPackageManifest      func(context.Context, string) (PackageRecord, error)
	ImportPackageBundle        func(context.Context, string) (PackageRecord, error)
	ApplyPackage               func(context.Context, string) (ActivePackageRecord, error)
	UpdateSetupField           SetupFieldUpdater
	InstallPackage             Installer
	Repair                     Repairer
	Uninstall                  Uninstaller
	RenderServiceHostArtifacts ServiceHostRenderer
	ValidateManifest           func([]byte) error
}

type Server struct {
	listen string
	deps   HandlerDependencies
}

func NewServer(listen string, deps HandlerDependencies) *Server {
	return &Server{listen: listen, deps: deps}
}

func (s *Server) Run(ctx context.Context) error {
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
	mux.HandleFunc("/api/v1/updates/apply", s.handleApplyPackage)
	mux.HandleFunc("/api/v1/setup/fields", s.handleSetupFieldUpdate)
	mux.HandleFunc("/api/v1/install", s.handleInstallPackage)
	mux.HandleFunc("/api/v1/repair", s.handleRepair)
	mux.HandleFunc("/api/v1/uninstall", s.handleUninstall)
	mux.HandleFunc("/api/v1/service-host/render", s.handleServiceHostRender)
	mux.HandleFunc("/api/v1/manifests/validate", s.handleManifestValidation)

	server := &http.Server{
		Addr:              s.listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]string{"status": "ok"}})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.deps.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": status})
}

func (s *Server) handleActivationRequest(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Customer string `json:"customer"`
		Output   string `json:"output"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	path, err := s.deps.GenerateActivationRequest(r.Context(), request.Customer, request.Output)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": map[string]string{"path": path}})
}

func (s *Server) handleLicenseImport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Path string `json:"path"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if request.Path == "" {
		writeError(w, http.StatusBadRequest, errors.New("path is required"))
		return
	}
	if err := s.deps.ImportLicense(r.Context(), request.Path); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": map[string]string{"status": "imported"}})
}

func (s *Server) handleManifestValidation(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Manifest json.RawMessage `json:"manifest"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.deps.ValidateManifest(request.Manifest); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true, "data": map[string]string{"status": "valid"}})
}

func (s *Server) handleServiceStart(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.deps.StartService(r.Context(), request.Name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": map[string]string{"status": "started"}})
}

func (s *Server) handleServiceStop(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.deps.StopService(request.Name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": map[string]string{"status": "stopped"}})
}

func (s *Server) handleServiceRestart(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Name string `json:"name"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.deps.RestartService(r.Context(), request.Name); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": map[string]string{"status": "restarted"}})
}

func (s *Server) handleManifestImport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Path string `json:"path"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	record, err := s.deps.ImportPackageManifest(r.Context(), request.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": record})
}

func (s *Server) handleBundleImport(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Path string `json:"path"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	record, err := s.deps.ImportPackageBundle(r.Context(), request.Path)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": record})
}

func (s *Server) handleApplyPackage(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PackageID string `json:"packageId"`
	}
	if err := decodeBody(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	record, err := s.deps.ApplyPackage(r.Context(), request.PackageID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"success": true, "data": record})
}

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
