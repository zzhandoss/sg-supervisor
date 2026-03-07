package packaging

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
)

type AssembleReport struct {
	Platform      string   `json:"platform"`
	OutputDir     string   `json:"outputDir"`
	ManifestPath  string   `json:"manifestPath"`
	CopiedEntries []string `json:"copiedEntries"`
}

func LoadManifest(path string) (Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return Manifest{}, err
	}
	return manifest, nil
}

func Assemble(root string, manifest Manifest) (AssembleReport, error) {
	outputDir := filepath.Join(root, "build", manifest.Platform)
	if err := os.RemoveAll(outputDir); err != nil {
		return AssembleReport{}, err
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return AssembleReport{}, err
	}

	copied := make([]string, 0, len(manifest.Files)+1)
	manifestPath := filepath.Join(outputDir, "install-manifest.json")
	if err := writeManifestFile(manifestPath, manifest); err != nil {
		return AssembleReport{}, err
	}
	copied = append(copied, manifestPath)

	for _, entry := range manifest.Files {
		targetPath := filepath.Join(outputDir, entry.TargetPath)
		if err := copyPath(entry.SourcePath, targetPath); err != nil {
			return AssembleReport{}, err
		}
		copied = append(copied, targetPath)
	}

	return AssembleReport{
		Platform:      manifest.Platform,
		OutputDir:     outputDir,
		ManifestPath:  manifestPath,
		CopiedEntries: copied,
	}, nil
}

func writeManifestFile(path string, manifest Manifest) error {
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func copyPath(sourcePath, targetPath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return copyFile(sourcePath, targetPath, info.Mode())
	}
	return filepath.WalkDir(sourcePath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}
		currentTarget := filepath.Join(targetPath, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(currentTarget, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return copyFile(path, currentTarget, info.Mode())
	})
}

func copyFile(sourcePath, targetPath string, mode os.FileMode) error {
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, data, mode)
}
