package control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSetupFieldUpdateHandler(t *testing.T) {
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
		UpdateSetupField: func(context.Context, string, string) (SetupStatus, error) {
			return SetupStatus{Complete: true}, nil
		},
		InstallPackage:             func(context.Context, string, string) (InstallReport, error) { return InstallReport{}, nil },
		Repair:                     func(context.Context, string) (RepairReport, error) { return RepairReport{}, nil },
		Uninstall:                  func(context.Context, string) (UninstallReport, error) { return UninstallReport{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/setup/fields", strings.NewReader(`{"key":"telegram-bot","status":"completed"}`))
	response := httptest.NewRecorder()

	server.handleSetupFieldUpdate(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
}
