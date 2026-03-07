package control

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestInstallPackageHandler(t *testing.T) {
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
		InstallPackage: func(context.Context, string, string) (InstallReport, error) {
			return InstallReport{ServiceName: "school-gate-supervisor"}, nil
		},
		Repair:                     func(context.Context, string) (RepairReport, error) { return RepairReport{}, nil },
		Uninstall:                  func(context.Context, string) (UninstallReport, error) { return UninstallReport{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/install", strings.NewReader(`{"packageId":"pkg-1","binaryPath":"C:\\svc\\sg-supervisor.exe"}`))
	response := httptest.NewRecorder()

	server.handleInstallPackage(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestRepairHandler(t *testing.T) {
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
		InstallPackage:            func(context.Context, string, string) (InstallReport, error) { return InstallReport{}, nil },
		Repair: func(context.Context, string) (RepairReport, error) {
			return RepairReport{NeedsPackageInstall: true}, nil
		},
		Uninstall:                  func(context.Context, string) (UninstallReport, error) { return UninstallReport{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/repair", strings.NewReader(`{"binaryPath":"C:\\svc\\sg-supervisor.exe"}`))
	response := httptest.NewRecorder()

	server.handleRepair(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestUninstallHandler(t *testing.T) {
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
		InstallPackage:            func(context.Context, string, string) (InstallReport, error) { return InstallReport{}, nil },
		Repair:                    func(context.Context, string) (RepairReport, error) { return RepairReport{}, nil },
		Uninstall: func(context.Context, string) (UninstallReport, error) {
			return UninstallReport{Mode: "keep-state"}, nil
		},
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/uninstall", strings.NewReader(`{"mode":"keep-state"}`))
	response := httptest.NewRecorder()

	server.handleUninstall(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}

func TestUninstallHandlerReturnsPartialDataOnFailure(t *testing.T) {
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
		InstallPackage:            func(context.Context, string, string) (InstallReport, error) { return InstallReport{}, nil },
		Repair:                    func(context.Context, string) (RepairReport, error) { return RepairReport{}, nil },
		Uninstall: func(context.Context, string) (UninstallReport, error) {
			return UninstallReport{
				Mode:      "keep-state",
				Completed: true,
				Issues:    []Issue{{Step: "service-deregistration", Message: "boom"}},
			}, errors.New("uninstall completed with issues")
		},
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/uninstall", strings.NewReader(`{"mode":"keep-state"}`))
	response := httptest.NewRecorder()

	server.handleUninstall(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", response.Code)
	}
	var body struct {
		Success bool            `json:"success"`
		Data    UninstallReport `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success {
		t.Fatalf("expected failure response")
	}
	if len(body.Data.Issues) != 1 {
		t.Fatalf("expected partial uninstall data")
	}
}
