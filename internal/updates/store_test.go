package updates

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"sg-supervisor/internal/config"
)

func TestImportManifest(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	sourcePath := filepath.Join(root, "package.json")
	manifestData := []byte(`{
  "productVersion":"1.2.0",
  "coreVersion":"1.2.0",
  "supervisorVersion":"0.1.0",
  "runtime":{"nodeVersion":"20.x"},
  "adapters":[{"key":"dahua-terminal-adapter","version":"0.2.0","required":true}],
  "compatibility":{"coreApi":1,"adapterApi":1}
}`)
	if err := os.WriteFile(sourcePath, manifestData, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(sourcePath+".sig", signManifest(t, privateKey, manifestData), 0o644); err != nil {
		t.Fatalf("write manifest signature: %v", err)
	}

	store := NewStore(layout, config.SupervisorConfig{
		PackageSigningPublicKeyBase64: base64.StdEncoding.EncodeToString(publicKey),
	})
	record, err := store.ImportManifest(context.Background(), sourcePath)
	if err != nil {
		t.Fatalf("import manifest: %v", err)
	}
	if record.PackageID == "" {
		t.Fatalf("expected package id")
	}

	records, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("list records: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected one record, got %d", len(records))
	}
}

func TestApplyManifest(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	sourcePath := filepath.Join(root, "package.json")
	manifestData := []byte(`{
  "productVersion":"1.2.0",
  "coreVersion":"1.2.0",
  "supervisorVersion":"0.1.0",
  "runtime":{"nodeVersion":"20.x"},
  "adapters":[{"key":"dahua-terminal-adapter","version":"0.2.0","required":true}],
  "compatibility":{"coreApi":1,"adapterApi":1}
}`)
	if err := os.WriteFile(sourcePath, manifestData, 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}
	if err := os.WriteFile(sourcePath+".sig", signManifest(t, privateKey, manifestData), 0o644); err != nil {
		t.Fatalf("write manifest signature: %v", err)
	}

	store := NewStore(layout, config.SupervisorConfig{
		PackageSigningPublicKeyBase64: base64.StdEncoding.EncodeToString(publicKey),
	})
	record, err := store.ImportManifest(context.Background(), sourcePath)
	if err != nil {
		t.Fatalf("import manifest: %v", err)
	}

	active, err := store.Apply(context.Background(), record.PackageID)
	if err != nil {
		t.Fatalf("apply manifest: %v", err)
	}
	if active.PackageID != record.PackageID {
		t.Fatalf("unexpected active record: %+v", active)
	}
}

func TestImportBundleAndApplyCopiesPayload(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	bundlePath := filepath.Join(root, "update-bundle.zip")
	manifestData := `{"productVersion":"1.2.0","coreVersion":"1.2.0","supervisorVersion":"0.1.0","runtime":{"nodeVersion":"20.x"},"adapters":[{"key":"dahua-terminal-adapter","version":"0.2.0","required":true}],"compatibility":{"coreApi":1,"adapterApi":1}}`
	if err := writeBundleZip(bundlePath, map[string]string{
		"manifest.json":                                   manifestData,
		"manifest.sig":                                    string(signManifest(t, privateKey, []byte(manifestData))),
		"payload/core/RELEASE_MANIFEST.json":              `{"version":"1.2.0"}`,
		"payload/adapters/dahua-terminal-adapter/VERSION": "0.2.0",
	}); err != nil {
		t.Fatalf("write bundle zip: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(layout.InstallDir, "core"), 0o755); err != nil {
		t.Fatalf("mkdir existing core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layout.InstallDir, "core", "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	store := NewStore(layout, config.SupervisorConfig{
		PackageSigningPublicKeyBase64: base64.StdEncoding.EncodeToString(publicKey),
	})
	record, err := store.ImportBundle(context.Background(), bundlePath)
	if err != nil {
		t.Fatalf("import bundle: %v", err)
	}
	if record.SourceType != "bundle" {
		t.Fatalf("expected bundle source type, got %s", record.SourceType)
	}

	active, err := store.Apply(context.Background(), record.PackageID)
	if err != nil {
		t.Fatalf("apply bundle: %v", err)
	}
	if active.PackageID != record.PackageID {
		t.Fatalf("unexpected active package: %+v", active)
	}

	if _, err := os.Stat(filepath.Join(layout.InstallDir, "core", "RELEASE_MANIFEST.json")); err != nil {
		t.Fatalf("expected copied core payload: %v", err)
	}
	if _, err := os.Stat(filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter", "VERSION")); err != nil {
		t.Fatalf("expected copied adapter payload: %v", err)
	}
	if _, err := os.Stat(filepath.Join(layout.BackupsDir, record.PackageID, "core", "old.txt")); err != nil {
		t.Fatalf("expected backed up old core payload: %v", err)
	}
}

func TestRollbackRestoresPreviousPayload(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	if err := os.MkdirAll(filepath.Join(layout.InstallDir, "core"), 0o755); err != nil {
		t.Fatalf("mkdir install core: %v", err)
	}
	if err := os.WriteFile(filepath.Join(layout.InstallDir, "core", "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("write old payload: %v", err)
	}

	bundlePath := filepath.Join(root, "rollback-bundle.zip")
	manifestData := `{"productVersion":"1.2.0","coreVersion":"1.2.0","supervisorVersion":"0.1.0","runtime":{"nodeVersion":"20.x"},"compatibility":{"coreApi":1,"adapterApi":1}}`
	if err := writeBundleZip(bundlePath, map[string]string{
		"manifest.json":        manifestData,
		"manifest.sig":         string(signManifest(t, privateKey, []byte(manifestData))),
		"payload/core/new.txt": "new",
		"payload/runtime/node": "runtime",
		"payload/adapters/x/a": "adapter",
	}); err != nil {
		t.Fatalf("write bundle zip: %v", err)
	}

	store := NewStore(layout, config.SupervisorConfig{
		PackageSigningPublicKeyBase64: base64.StdEncoding.EncodeToString(publicKey),
	})
	record, err := store.ImportBundle(context.Background(), bundlePath)
	if err != nil {
		t.Fatalf("import bundle: %v", err)
	}
	active, err := store.Apply(context.Background(), record.PackageID)
	if err != nil {
		t.Fatalf("apply bundle: %v", err)
	}
	if err := store.Rollback(context.Background(), active.BackupPath); err != nil {
		t.Fatalf("rollback bundle: %v", err)
	}

	if _, err := os.Stat(filepath.Join(layout.InstallDir, "core", "old.txt")); err != nil {
		t.Fatalf("expected restored old payload: %v", err)
	}
	if _, err := os.Stat(filepath.Join(layout.InstallDir, "core", "new.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected new payload removed on rollback, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(layout.InstallDir, "runtime")); !os.IsNotExist(err) {
		t.Fatalf("expected created runtime payload removed on rollback, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(layout.InstallDir, "adapters", "x")); !os.IsNotExist(err) {
		t.Fatalf("expected created adapter payload removed on rollback, got %v", err)
	}
}

func TestImportBundleRejectsInvalidSignature(t *testing.T) {
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	bundlePath := filepath.Join(root, "invalid-bundle.zip")
	if err := writeBundleZip(bundlePath, map[string]string{
		"manifest.json":  `{"productVersion":"1.2.0","coreVersion":"1.2.0","supervisorVersion":"0.1.0","runtime":{"nodeVersion":"20.x"},"compatibility":{"coreApi":1,"adapterApi":1}}`,
		"manifest.sig":   base64.StdEncoding.EncodeToString([]byte("invalid")),
		"payload/core/x": "broken",
	}); err != nil {
		t.Fatalf("write bundle zip: %v", err)
	}

	store := NewStore(layout, config.SupervisorConfig{
		PackageSigningPublicKeyBase64: base64.StdEncoding.EncodeToString(publicKey),
	})
	if _, err := store.ImportBundle(context.Background(), bundlePath); err == nil {
		t.Fatalf("expected signature verification error")
	}
}

func TestSaveAndLoadOperationStatus(t *testing.T) {
	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}

	store := NewStore(layout, config.SupervisorConfig{})
	expected := OperationStatus{
		Action:          "apply-package",
		PackageID:       "pkg-1",
		Outcome:         "rolled-back",
		RollbackOutcome: "succeeded",
		ActivePackageID: "pkg-0",
		Message:         "restart failed",
	}
	if err := store.SaveOperation(expected); err != nil {
		t.Fatalf("save operation: %v", err)
	}

	actual, err := store.Operation(context.Background())
	if err != nil {
		t.Fatalf("load operation: %v", err)
	}
	if actual.Action != expected.Action || actual.PackageID != expected.PackageID || actual.Outcome != expected.Outcome {
		t.Fatalf("unexpected operation status: %+v", actual)
	}
	if actual.RecordedAt == "" {
		t.Fatalf("expected recordedAt to be populated")
	}
}

func writeBundleZip(path string, files map[string]string) error {
	buffer := &bytes.Buffer{}
	writer := zip.NewWriter(buffer)
	for name, content := range files {
		entry, err := writer.Create(name)
		if err != nil {
			return err
		}
		if _, err := entry.Write([]byte(content)); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buffer.Bytes(), 0o644)
}

func signManifest(t *testing.T, privateKey ed25519.PrivateKey, data []byte) []byte {
	t.Helper()
	signature := ed25519.Sign(privateKey, data)
	return []byte(base64.StdEncoding.EncodeToString(signature))
}
