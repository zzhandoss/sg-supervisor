package control

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"sg-supervisor/internal/license"
)

func TestStatusHandler(t *testing.T) {
	server := NewServer("127.0.0.1:0", HandlerDependencies{
		Status: func(context.Context) (StatusResponse, error) {
			return StatusResponse{
				ProductName:   "School Gate",
				SetupRequired: true,
				License:       license.Status{Valid: false},
			}, nil
		},
		GenerateActivationRequest:  func(context.Context, string, string) (string, error) { return "", nil },
		ImportLicense:              func(context.Context, string) error { return nil },
		StartService:               func(context.Context, string) error { return nil },
		StopService:                func(string) error { return nil },
		RestartService:             func(context.Context, string) error { return nil },
		ImportPackageManifest:      func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ImportPackageBundle:        func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ApplyPackage:               func(context.Context, string) (ActivePackageRecord, error) { return ActivePackageRecord{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodGet, "/api/v1/status", nil)
	response := httptest.NewRecorder()

	server.handleStatus(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}

	var body struct {
		Success bool           `json:"success"`
		Data    StatusResponse `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success {
		t.Fatalf("expected success response")
	}
	if body.Data.ProductName != "School Gate" {
		t.Fatalf("unexpected product name: %s", body.Data.ProductName)
	}
}

func TestManifestValidationHandler(t *testing.T) {
	server := NewServer("127.0.0.1:0", HandlerDependencies{
		Status:                     func(context.Context) (StatusResponse, error) { return StatusResponse{}, nil },
		GenerateActivationRequest:  func(context.Context, string, string) (string, error) { return "", nil },
		ImportLicense:              func(context.Context, string) error { return nil },
		StartService:               func(context.Context, string) error { return nil },
		StopService:                func(string) error { return nil },
		RestartService:             func(context.Context, string) error { return nil },
		ImportPackageManifest:      func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ImportPackageBundle:        func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ApplyPackage:               func(context.Context, string) (ActivePackageRecord, error) { return ActivePackageRecord{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest: func(data []byte) error {
			if !strings.Contains(string(data), "productVersion") {
				t.Fatalf("expected manifest payload")
			}
			return nil
		},
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/manifests/validate", strings.NewReader(`{"manifest":{"productVersion":"1.0.0","coreVersion":"1.0.0","supervisorVersion":"0.1.0","runtime":{"nodeVersion":"20.x"},"compatibility":{"coreApi":1,"adapterApi":1}}}`))
	response := httptest.NewRecorder()

	server.handleManifestValidation(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
}

func TestManifestImportHandler(t *testing.T) {
	server := NewServer("127.0.0.1:0", HandlerDependencies{
		Status:                    func(context.Context) (StatusResponse, error) { return StatusResponse{}, nil },
		GenerateActivationRequest: func(context.Context, string, string) (string, error) { return "", nil },
		ImportLicense:             func(context.Context, string) error { return nil },
		StartService:              func(context.Context, string) error { return nil },
		StopService:               func(string) error { return nil },
		RestartService:            func(context.Context, string) error { return nil },
		ImportPackageManifest: func(context.Context, string) (PackageRecord, error) {
			return PackageRecord{PackageID: "pkg-1"}, nil
		},
		ImportPackageBundle: func(context.Context, string) (PackageRecord, error) {
			return PackageRecord{PackageID: "bundle-1"}, nil
		},
		ApplyPackage:               func(context.Context, string) (ActivePackageRecord, error) { return ActivePackageRecord{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/updates/import-manifest", strings.NewReader(`{"path":"C:\\temp\\manifest.json"}`))
	response := httptest.NewRecorder()

	server.handleManifestImport(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestBundleImportHandler(t *testing.T) {
	server := NewServer("127.0.0.1:0", HandlerDependencies{
		Status:                    func(context.Context) (StatusResponse, error) { return StatusResponse{}, nil },
		GenerateActivationRequest: func(context.Context, string, string) (string, error) { return "", nil },
		ImportLicense:             func(context.Context, string) error { return nil },
		StartService:              func(context.Context, string) error { return nil },
		StopService:               func(string) error { return nil },
		RestartService:            func(context.Context, string) error { return nil },
		ImportPackageManifest:     func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ImportPackageBundle: func(context.Context, string) (PackageRecord, error) {
			return PackageRecord{PackageID: "bundle-1"}, nil
		},
		ApplyPackage:               func(context.Context, string) (ActivePackageRecord, error) { return ActivePackageRecord{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/updates/import-bundle", strings.NewReader(`{"path":"C:\\temp\\bundle.zip"}`))
	response := httptest.NewRecorder()

	server.handleBundleImport(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestApplyPackageHandler(t *testing.T) {
	server := NewServer("127.0.0.1:0", HandlerDependencies{
		Status:                    func(context.Context) (StatusResponse, error) { return StatusResponse{}, nil },
		GenerateActivationRequest: func(context.Context, string, string) (string, error) { return "", nil },
		ImportLicense:             func(context.Context, string) error { return nil },
		StartService:              func(context.Context, string) error { return nil },
		StopService:               func(string) error { return nil },
		RestartService:            func(context.Context, string) error { return nil },
		ImportPackageManifest:     func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ImportPackageBundle:       func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ApplyPackage: func(context.Context, string) (ActivePackageRecord, error) {
			return ActivePackageRecord{PackageID: "pkg-1"}, nil
		},
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/updates/apply", strings.NewReader(`{"packageId":"pkg-1"}`))
	response := httptest.NewRecorder()

	server.handleApplyPackage(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestServiceHostRenderHandler(t *testing.T) {
	server := NewServer("127.0.0.1:0", HandlerDependencies{
		Status:                    func(context.Context) (StatusResponse, error) { return StatusResponse{}, nil },
		GenerateActivationRequest: func(context.Context, string, string) (string, error) { return "", nil },
		ImportLicense:             func(context.Context, string) error { return nil },
		StartService:              func(context.Context, string) error { return nil },
		StopService:               func(string) error { return nil },
		RestartService:            func(context.Context, string) error { return nil },
		ImportPackageManifest:     func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ImportPackageBundle:       func(context.Context, string) (PackageRecord, error) { return PackageRecord{}, nil },
		ApplyPackage:              func(context.Context, string) (ActivePackageRecord, error) { return ActivePackageRecord{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) {
			return ServiceHostArtifacts{ServiceName: "school-gate-supervisor"}, nil
		},
		ValidateManifest: func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/service-host/render", strings.NewReader(`{"binaryPath":"C:\\svc\\sg-supervisor.exe"}`))
	response := httptest.NewRecorder()

	server.handleServiceHostRender(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}
