package releasepanel

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sg-supervisor/internal/config"
)

func TestServiceStatusIncludesGeneratedKeys(t *testing.T) {
	service, err := NewService(t.TempDir(), ".")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	status, err := service.Status(context.Background())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Keys.LicenseConfigured || !status.Keys.PackageConfigured {
		t.Fatalf("expected signing keys to be generated, got %+v", status.Keys)
	}
	if status.HostPlatform == "" {
		t.Fatal("expected host platform in status")
	}
}

func TestIssueLicenseFromActivationRequest(t *testing.T) {
	root := t.TempDir()
	service, err := NewService(root, ".")
	if err != nil {
		t.Fatalf("new service: %v", err)
	}
	requestPath := filepath.Join(root, "activation-request.json")
	data := []byte("{\n  \"product\": \"School Gate\",\n  \"requestedAt\": \"2026-03-08T00:00:00Z\",\n  \"customerHint\": \"Acme\",\n  \"fingerprint\": \"fp-123\",\n  \"signals\": [\"sig\"]\n}\n")
	if err := os.WriteFile(requestPath, data, 0o644); err != nil {
		t.Fatalf("write request: %v", err)
	}
	record, err := service.IssueLicense(context.Background(), LicenseIssueRequest{
		ActivationRequestPath: requestPath,
		Mode:                  "bound",
		Edition:               "standard",
		Perpetual:             true,
	})
	if err != nil {
		t.Fatalf("issue license: %v", err)
	}
	if record.Customer != "Acme" || record.Fingerprint != "fp-123" {
		t.Fatalf("unexpected record: %+v", record)
	}
	if _, err := os.Stat(record.Path); err != nil {
		t.Fatalf("expected license file: %v", err)
	}
}

func TestBuildLocalReleaseCreatesOwnerArtifacts(t *testing.T) {
	root := t.TempDir()
	layout := NewLayout(root)
	store := NewStore(layout)
	state, err := store.Ensure(".")
	if err != nil {
		t.Fatalf("ensure store: %v", err)
	}
	state.Recipe = Recipe{
		InstallerVersion:  "1.0.0",
		SchoolGateVersion: "1.2.0",
		AdapterVersion:    "0.2.0",
		NodeVersion:       "20.19.0",
	}
	if err := store.Save(state); err != nil {
		t.Fatalf("save state: %v", err)
	}
	service := &Service{
		layout:  layout,
		store:   store,
		jobs:    NewJobStore(layout),
		assets:  newFakeAssetSource(t),
		node:    newFakeNodeSource(t),
		core:    fakeCoreBuilder{},
		builder: fakeBinaryBuilder{},
	}
	job := Job{ID: "job-1", Type: JobTypeLocalRelease, Status: JobStatusRunning, Recipe: state.Recipe}
	report, err := service.buildLocalRelease(context.Background(), &job, state)
	if err != nil {
		t.Fatalf("build local release: %v", err)
	}
	if len(report.Reports) != 1 {
		t.Fatalf("expected one platform report, got %+v", report)
	}
	if len(report.Platforms) != 1 {
		t.Fatalf("expected one platform entry, got %+v", report.Platforms)
	}
	if _, err := os.Stat(filepath.Join(layout.ReleasesDir, "v1.0.0", "release-set.json")); err != nil {
		t.Fatalf("expected release-set metadata: %v", err)
	}
}

type fakeBinaryBuilder struct{}

func (fakeBinaryBuilder) BuildSupervisor(_ context.Context, _ string, _ string, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte("binary"), 0o755)
}

type fakeAssetSource struct {
	assets map[string]string
}

func newFakeAssetSource(t *testing.T) *fakeAssetSource {
	t.Helper()
	root := t.TempDir()
	return &fakeAssetSource{
		assets: map[string]string{
			"dahua-adapter-v0.2.0-win-x64.zip":   writeZipArchive(t, filepath.Join(root, "adapter-win.zip"), adapterFiles()),
			"dahua-adapter-v0.2.0-linux-x64.zip": writeZipArchive(t, filepath.Join(root, "adapter-linux.zip"), adapterFiles()),
		},
	}
}

func (s *fakeAssetSource) ListVersions(_ context.Context, _ string) ([]ReleaseVersion, error) {
	return nil, nil
}

func (s *fakeAssetSource) DownloadReleaseAsset(_ context.Context, spec AssetSpec, _ string) (string, error) {
	return s.assets[spec.Pattern], nil
}

type fakeNodeSource struct {
	assets map[string]string
}

type fakeCoreBuilder struct{}

func (fakeCoreBuilder) BuildInstallTree(_ context.Context, _ Recipe, _ string, workspaceRoot string, _ func(string)) error {
	layout := config.NewLayout(workspaceRoot)
	for path, body := range coreFiles() {
		trimmed := strings.TrimPrefix(path, "school-gate-v1.2.0/")
		target := filepath.Join(layout.InstallDir, "core", filepath.FromSlash(trimmed))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(body), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func newFakeNodeSource(t *testing.T) *fakeNodeSource {
	t.Helper()
	root := t.TempDir()
	return &fakeNodeSource{
		assets: map[string]string{
			"windows": writeZipArchive(t, filepath.Join(root, "node-win.zip"), map[string]string{"node-v20.19.0-win-x64/node.exe": "node"}),
			"linux":   writeTarGzArchive(t, filepath.Join(root, "node-linux.tar.gz"), map[string]string{"node-v20.19.0-linux-x64/bin/node": "node"}),
		},
	}
}

func (s *fakeNodeSource) ListVersions(_ context.Context) ([]ReleaseVersion, error) {
	return nil, nil
}

func (s *fakeNodeSource) Download(_ string, platform, _ string) (string, error) {
	return s.assets[platform], nil
}

func writeZipArchive(t *testing.T, path string, files map[string]string) string {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create zip: %v", err)
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	for name, body := range files {
		record, err := writer.Create(name)
		if err != nil {
			t.Fatalf("create record: %v", err)
		}
		if _, err := record.Write([]byte(body)); err != nil {
			t.Fatalf("write record: %v", err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return path
}

func writeTarGzArchive(t *testing.T, path string, files map[string]string) string {
	t.Helper()
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create tar.gz: %v", err)
	}
	defer file.Close()
	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()
	for name, body := range files {
		header := &tar.Header{Name: name, Mode: 0o755, Size: int64(len(body))}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(body)); err != nil {
			t.Fatalf("write body: %v", err)
		}
	}
	return path
}
