package updates

import (
	"archive/zip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sg-supervisor/internal/manifest"
)

func (s *Store) ImportBundle(ctx context.Context, sourcePath string) (Record, error) {
	if err := ctx.Err(); err != nil {
		return Record{}, err
	}
	if err := s.ensure(); err != nil {
		return Record{}, err
	}

	packageID := time.Now().UTC().Format("20060102T150405Z")
	stageDir := filepath.Join(s.stagingDir(), packageID)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return Record{}, err
	}

	archivePath := filepath.Join(s.packagesDir(), packageID+filepath.Ext(sourcePath))
	if err := copyFile(sourcePath, archivePath); err != nil {
		return Record{}, err
	}

	storedManifestPath := filepath.Join(s.manifestsDir(), packageID+".json")
	file, err := s.extractBundle(archivePath, stageDir, storedManifestPath)
	if err != nil {
		return Record{}, err
	}

	record := Record{
		PackageID:   packageID,
		SourceType:  "bundle",
		SourcePath:  sourcePath,
		ArchivePath: archivePath,
		StoredPath:  storedManifestPath,
		StageDir:    stageDir,
		ImportedAt:  time.Now().UTC().Format(time.RFC3339),
		Manifest:    file,
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

func (s *Store) extractBundle(archivePath, stageDir, storedManifestPath string) (manifest.File, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return manifest.File{}, err
	}
	defer reader.Close()

	var manifestFound bool
	var manifestData []byte
	var signatureData []byte
	for _, file := range reader.File {
		entryData, err := extractZipEntry(file, stageDir, storedManifestPath)
		if err != nil {
			return manifest.File{}, err
		}
		switch cleanZipPath(file.Name) {
		case "manifest.json":
			manifestFound = true
			manifestData = entryData
		case "manifest.sig":
			signatureData = entryData
		}
	}
	if !manifestFound {
		return manifest.File{}, errors.New("bundle does not contain manifest.json")
	}
	if len(signatureData) == 0 {
		return manifest.File{}, errors.New("bundle does not contain manifest.sig")
	}
	if err := verifyManifestSignature(s.cfg, manifestData, signatureData); err != nil {
		return manifest.File{}, err
	}

	data, err := os.ReadFile(storedManifestPath)
	if err != nil {
		return manifest.File{}, err
	}
	var bundleManifest manifest.File
	if err := json.Unmarshal(data, &bundleManifest); err != nil {
		return manifest.File{}, err
	}
	if err := manifest.Validate(bundleManifest); err != nil {
		return manifest.File{}, err
	}
	return bundleManifest, nil
}

func extractZipEntry(file *zip.File, stageDir, storedManifestPath string) ([]byte, error) {
	name := cleanZipPath(file.Name)
	if name == "" {
		return nil, nil
	}

	if name == "manifest.json" {
		return extractFileToPath(file, storedManifestPath)
	}
	if name == "manifest.sig" {
		return readZipEntry(file)
	}

	targetPath, ok := safeJoin(stageDir, name)
	if !ok {
		return nil, errors.New("bundle contains unsafe path")
	}

	if file.FileInfo().IsDir() {
		return nil, os.MkdirAll(targetPath, 0o755)
	}
	return extractFileToPath(file, targetPath)
}

func extractFileToPath(file *zip.File, targetPath string) ([]byte, error) {
	data, err := readZipEntry(file)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return nil, err
	}
	target, err := os.Create(targetPath)
	if err != nil {
		return nil, err
	}
	defer target.Close()

	if _, err := target.Write(data); err != nil {
		return nil, err
	}
	return data, nil
}

func cleanZipPath(path string) string {
	path = filepath.ToSlash(path)
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "/")
	return filepath.Clean(path)
}

func safeJoin(base, name string) (string, bool) {
	target := filepath.Join(base, name)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", false
	}
	if strings.HasPrefix(rel, "..") {
		return "", false
	}
	return target, true
}

func copyFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	target, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)
	return err
}

func readZipEntry(file *zip.File) ([]byte, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}
