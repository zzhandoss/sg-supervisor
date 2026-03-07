package app

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/servicehost"
)

type recordingRunner struct {
	actions []servicehost.Action
	failAt  string
}

func (r *recordingRunner) Run(_ context.Context, action servicehost.Action) error {
	r.actions = append(r.actions, action)
	if action.Name == r.failAt {
		return errors.New("boom")
	}
	return nil
}

func TestRepairExecutesServiceHostActions(t *testing.T) {
	supervisor := newTestApp(t)
	runner := &recordingRunner{}
	supervisor.runner = runner

	report, err := supervisor.Repair(context.Background(), filepath.Join(supervisor.root, "sg-supervisor.exe"))
	if err != nil {
		t.Fatalf("repair: %v", err)
	}
	if len(report.ServiceArtifacts) == 0 {
		t.Fatalf("expected rendered artifacts")
	}
	if len(runner.actions) == 0 {
		t.Fatalf("expected repair to execute service-host actions")
	}
	assertPackagingManifest(t, supervisor.root, "linux")
	assertPackagingManifest(t, supervisor.root, "windows")
}

func TestInstallPackageExecutesServiceHostActions(t *testing.T) {
	supervisor, privateKey := newSignedTestApp(t)
	runner := &recordingRunner{}
	supervisor.runner = runner

	manifestPath := filepath.Join(supervisor.root, "package.json")
	manifestData := []byte(`{
  "productVersion":"1.2.0",
  "coreVersion":"1.2.0",
  "supervisorVersion":"0.1.0",
  "runtime":{"nodeVersion":"20.x"},
  "compatibility":{"coreApi":1,"adapterApi":1}
}`)
	if err := os.WriteFile(manifestPath, manifestData, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	signature := ed25519.Sign(privateKey, manifestData)
	if err := os.WriteFile(manifestPath+".sig", []byte(base64.StdEncoding.EncodeToString(signature)), 0o644); err != nil {
		t.Fatalf("write signature: %v", err)
	}

	record, err := supervisor.ImportPackageManifest(context.Background(), manifestPath)
	if err != nil {
		t.Fatalf("import manifest: %v", err)
	}
	report, err := supervisor.InstallPackage(context.Background(), record.PackageID, filepath.Join(supervisor.root, "sg-supervisor.exe"))
	if err != nil {
		t.Fatalf("install package: %v", err)
	}
	if report.ActivePackageID == "" {
		t.Fatalf("expected active package id")
	}
	if len(runner.actions) == 0 {
		t.Fatalf("expected install to execute service-host actions")
	}
	assertPackagingManifest(t, supervisor.root, "linux")
	assertPackagingManifest(t, supervisor.root, "windows")
}

func TestUninstallExecutesServiceHostActions(t *testing.T) {
	supervisor := newTestApp(t)
	runner := &recordingRunner{}
	supervisor.runner = runner

	report, err := supervisor.Uninstall(context.Background(), "keep-state")
	if err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if len(report.RemovedPaths) == 0 {
		t.Fatalf("expected removed paths")
	}
	if len(runner.actions) == 0 {
		t.Fatalf("expected uninstall to execute service-host actions")
	}
}

func TestUninstallReturnsPartialReportOnDeregistrationFailure(t *testing.T) {
	supervisor := newTestApp(t)
	failAt := "disable-service"
	if runtime.GOOS == "windows" {
		failAt = "uninstall-service"
	}
	runner := &recordingRunner{failAt: failAt}
	supervisor.runner = runner

	report, err := supervisor.Uninstall(context.Background(), "keep-state")
	if err == nil {
		t.Fatalf("expected uninstall issues to surface as error")
	}
	if !report.Completed {
		t.Fatalf("expected filesystem uninstall to complete")
	}
	if len(report.Issues) == 0 {
		t.Fatalf("expected uninstall issues in report")
	}
}

func TestAssemblePackageCreatesBuildOutput(t *testing.T) {
	supervisor := newTestApp(t)
	runner := &recordingRunner{}
	supervisor.runner = runner

	binaryPath := filepath.Join(supervisor.root, "sg-supervisor.exe")
	if err := os.WriteFile(binaryPath, []byte("bin"), 0o755); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	if err := os.WriteFile(filepath.Join(supervisor.layout.InstallDir, "app.txt"), []byte("app"), 0o644); err != nil {
		t.Fatalf("write install file: %v", err)
	}

	report, err := supervisor.AssemblePackage(context.Background(), "windows", binaryPath)
	if err != nil {
		t.Fatalf("assemble package: %v", err)
	}
	if _, err := os.Stat(filepath.Join(report.OutputDir, "install", "app.txt")); err != nil {
		t.Fatalf("expected install file in assembled output: %v", err)
	}
}

func newTestApp(t *testing.T) *App {
	t.Helper()
	root := t.TempDir()
	supervisor, err := New(root)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	return supervisor
}

func newSignedTestApp(t *testing.T) (*App, ed25519.PrivateKey) {
	t.Helper()
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	data := []byte("{\n  \"productName\": \"School Gate\",\n  \"listenAddress\": \"127.0.0.1:8787\",\n  \"packageSigningPublicKeyBase64\": \"" + base64.StdEncoding.EncodeToString(publicKey) + "\"\n}\n")
	if err := os.WriteFile(layout.ConfigFile, data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	supervisor, err := New(root)
	if err != nil {
		t.Fatalf("new app: %v", err)
	}
	return supervisor, privateKey
}

func assertPackagingManifest(t *testing.T, root, platform string) {
	t.Helper()
	path := filepath.Join(root, "runtime", "packaging", platform, "install-manifest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read packaging manifest: %v", err)
	}
	var body struct {
		Platform        string `json:"platform"`
		ActivePackageID string `json:"activePackageId"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatalf("decode packaging manifest: %v", err)
	}
	if body.Platform != platform {
		t.Fatalf("unexpected platform in manifest: %+v", body)
	}
}
