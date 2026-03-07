package updates

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/manifest"
)

type Record struct {
	PackageID   string        `json:"packageId"`
	SourceType  string        `json:"sourceType"`
	SourcePath  string        `json:"sourcePath"`
	ArchivePath string        `json:"archivePath,omitempty"`
	StoredPath  string        `json:"storedPath"`
	StageDir    string        `json:"stageDir,omitempty"`
	ImportedAt  string        `json:"importedAt"`
	Manifest    manifest.File `json:"manifest"`
}

type Store struct {
	layout config.Layout
	cfg    config.SupervisorConfig
}

func NewStore(layout config.Layout, cfg config.SupervisorConfig) *Store {
	return &Store{layout: layout, cfg: cfg}
}

func (s *Store) ImportManifest(ctx context.Context, sourcePath string) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	if err := s.ensure(); err != nil {
		return Record{}, err
	}

	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return Record{}, err
	}
	signatureData, err := readSignatureFile(sourcePath + ".sig")
	if err != nil {
		return Record{}, err
	}
	if err := verifyManifestSignature(s.cfg, data, signatureData); err != nil {
		return Record{}, err
	}

	var file manifest.File
	if err := json.Unmarshal(data, &file); err != nil {
		return Record{}, err
	}
	if err := manifest.Validate(file); err != nil {
		return Record{}, err
	}

	packageID := time.Now().UTC().Format("20060102T150405Z")
	storedPath := filepath.Join(s.manifestsDir(), packageID+".json")
	if err := os.WriteFile(storedPath, data, 0o644); err != nil {
		return Record{}, err
	}

	record := Record{
		PackageID:  packageID,
		SourceType: "manifest",
		SourcePath: sourcePath,
		StoredPath: storedPath,
		ImportedAt: time.Now().UTC().Format(time.RFC3339),
		Manifest:   file,
	}

	records, err := s.List(ctx)
	if err != nil {
		return Record{}, err
	}
	records = append(records, record)
	if err := s.writeIndex(records); err != nil {
		return Record{}, err
	}
	return record, nil
}

func (s *Store) List(ctx context.Context) ([]Record, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.ensure(); err != nil {
		return nil, err
	}
	data, err := os.ReadFile(s.indexPath())
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var records []Record
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func (s *Store) ensure() error {
	dirs := []string{
		s.manifestsDir(),
		s.packagesDir(),
		s.stagingDir(),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) manifestsDir() string {
	return filepath.Join(s.layout.UpdatesDir, "manifests")
}

func (s *Store) indexPath() string {
	return filepath.Join(s.layout.UpdatesDir, "imported-manifests.json")
}

func (s *Store) packagesDir() string {
	return filepath.Join(s.layout.UpdatesDir, "packages")
}

func (s *Store) stagingDir() string {
	return filepath.Join(s.layout.UpdatesDir, "staging")
}

func (s *Store) writeIndex(records []Record) error {
	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.indexPath(), data, 0o644)
}
