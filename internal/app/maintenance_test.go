package app

import (
	"archive/zip"
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
	"sg-supervisor/internal/manifest"
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

func TestBootstrapInstallAppliesLocalBundleAndRegistersService(t *testing.T) {
	supervisor, privateKey := newSignedTestApp(t)
	runner := &recordingRunner{}
	supervisor.runner = runner

	bundleDir := filepath.Join(supervisor.root, "delivery", "payload")
	if err := os.MkdirAll(bundleDir, 0o755); err != nil {
		t.Fatalf("mkdir bundle dir: %v", err)
	}

	layout := config.NewLayout(supervisor.root)
	if err := os.MkdirAll(filepath.Join(layout.InstallDir, "core", "apps", "api", "dist"), 0o755); err != nil {
		t.Fatalf("mkdir core: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", "dist", "src"), 0o755); err != nil {
		t.Fatalf("mkdir adapter: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layout.InstallDir, "core", "apps", "api", "dist", "index.js"), []byte("api"), 0o644); err != nil {
		t.Fatalf("write core file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", "dist", "src", "index.js"), []byte("adapter"), 0o644); err != nil {
		t.Fatalf("write adapter file: %v", err)
	}

	bundlePath := filepath.Join(bundleDir, "school-gate-package-v1.0.0-windows-x64.zip")
	if err := writeSignedBundle(bundlePath, privateKey, map[string]string{
		"payload/core/apps/api/dist/index.js":                       "api",
		"payload/adapters/dahua-terminal-adapter/dist/src/index.js": "adapter",
	}); err != nil {
		t.Fatalf("build payload bundle: %v", err)
	}

	report, err := supervisor.BootstrapInstall(context.Background(), bundleDir, filepath.Join(supervisor.root, "sg-supervisor.exe"))
	if err != nil {
		t.Fatalf("bootstrap install: %v", err)
	}
	if report.ActivePackageID == "" {
		t.Fatalf("expected active package id")
	}
	if len(runner.actions) == 0 {
		t.Fatalf("expected bootstrap install to execute service-host actions")
	}
}

func TestInstallPackageReturnsPartialReportOnServiceRegistrationFailure(t *testing.T) {
	supervisor, privateKey := newSignedTestApp(t)
	failAt := "enable-service"
	if runtime.GOOS == "windows" {
		failAt = "install-service"
	}
	runner := &recordingRunner{failAt: failAt}
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
	if err == nil {
		t.Fatalf("expected install issues to surface as error")
	}
	if report.ActivePackageID == "" || len(report.Issues) != 1 {
		t.Fatalf("expected partial install report")
	}
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

func TestRepairReturnsPartialReportOnServiceRepairFailure(t *testing.T) {
	supervisor := newTestApp(t)
	failAt := "enable-service"
	if runtime.GOOS == "windows" {
		failAt = "install-service"
	}
	runner := &recordingRunner{failAt: failAt}
	supervisor.runner = runner

	report, err := supervisor.Repair(context.Background(), filepath.Join(supervisor.root, "sg-supervisor.exe"))
	if err == nil {
		t.Fatalf("expected repair issues to surface as error")
	}
	if len(report.EnsuredPaths) == 0 || len(report.Issues) != 1 {
		t.Fatalf("expected partial repair report")
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

func writeSignedBundle(path string, privateKey ed25519.PrivateKey, payloadFiles map[string]string) error {
	file := manifest.File{
		ProductVersion:    "1.0.0",
		CoreVersion:       "1.2.0",
		SupervisorVersion: "0.1.0",
		Runtime: manifest.Runtime{
			NodeVersion: "20.19.0",
		},
		Adapters: []manifest.AdapterBundle{
			{Key: "dahua-terminal-adapter", Version: "0.2.0", Required: true},
		},
		Compatibility: manifest.Compatibility{
			CoreAPI:    1,
			AdapterAPI: 1,
		},
	}
	manifestData, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	manifestData = append(manifestData, '\n')
	signature := ed25519.Sign(privateKey, manifestData)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	archive, err := os.Create(path)
	if err != nil {
		return err
	}
	defer archive.Close()
	writer := zip.NewWriter(archive)
	defer writer.Close()
	if err := writeBundleZipEntry(writer, "manifest.json", manifestData); err != nil {
		return err
	}
	if err := writeBundleZipEntry(writer, "manifest.sig", []byte(base64.StdEncoding.EncodeToString(signature)+"\n")); err != nil {
		return err
	}
	for name, body := range payloadFiles {
		if err := writeBundleZipEntry(writer, name, []byte(body)); err != nil {
			return err
		}
	}
	return nil
}

func writeBundleZipEntry(writer *zip.Writer, name string, data []byte) error {
	record, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = record.Write(data)
	return err
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
