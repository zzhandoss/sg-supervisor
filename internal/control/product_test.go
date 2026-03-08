package control

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProductConfigUpdateHandler(t *testing.T) {
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
		UpdateSetupField:          func(context.Context, string, string, string) (SetupStatus, error) { return SetupStatus{}, nil },
		UpdateProductConfig: func(_ context.Context, update ProductConfigUpdate) (ProductConfigStatus, error) {
			if update.PreferredHost == nil || *update.PreferredHost != "10.20.30.40" {
				t.Fatalf("unexpected preferred host: %+v", update)
			}
			if update.TelegramBotToken == nil || *update.TelegramBotToken != "token-123" {
				t.Fatalf("unexpected telegram bot token: %+v", update)
			}
			return ProductConfigStatus{ResolvedHost: "10.20.30.40", TelegramBotConfigured: true, AdminUIURL: "http://10.20.30.40:5000"}, nil
		},
		InstallPackage:             func(context.Context, string, string) (InstallReport, error) { return InstallReport{}, nil },
		Repair:                     func(context.Context, string) (RepairReport, error) { return RepairReport{}, nil },
		Uninstall:                  func(context.Context, string) (UninstallReport, error) { return UninstallReport{}, nil },
		RenderServiceHostArtifacts: func(context.Context, string) (ServiceHostArtifacts, error) { return ServiceHostArtifacts{}, nil },
		ValidateManifest:           func([]byte) error { return nil },
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/product-config", strings.NewReader(`{"preferredHost":"10.20.30.40","telegramBotToken":"token-123"}`))
	response := httptest.NewRecorder()

	server.handleProductConfigUpdate(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", response.Code)
	}
	var body struct {
		Success bool                `json:"success"`
		Data    ProductConfigStatus `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Data.ResolvedHost != "10.20.30.40" || body.Data.AdminUIURL != "http://10.20.30.40:5000" {
		t.Fatalf("unexpected response: %+v", body)
	}
}
