package release

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sg-supervisor/internal/distribution"
)

func Build(root, version string, report distribution.Report) (Report, error) {
	version = normalizeVersion(version)
	releaseDir := filepath.Join(root, "releases", "v"+version, report.Platform)
	if err := os.RemoveAll(releaseDir); err != nil {
		return Report{}, err
	}
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		return Report{}, err
	}

	artifactPath, warnings, err := materializeArtifact(releaseDir, version, report)
	if err != nil {
		return Report{}, err
	}
	checksumsPath, err := writeChecksums(releaseDir, version, report.Platform, artifactPath)
	if err != nil {
		return Report{}, err
	}
	metadataPath, err := writeMetadata(releaseDir, version, report, artifactPath, checksumsPath, warnings)
	if err != nil {
		return Report{}, err
	}

	return Report{
		Version:       version,
		Platform:      report.Platform,
		ReleaseDir:    releaseDir,
		ArtifactPath:  artifactPath,
		MetadataPath:  metadataPath,
		ChecksumsPath: checksumsPath,
		Warnings:      warnings,
	}, nil
}

func materializeArtifact(releaseDir, version string, report distribution.Report) (string, []string, error) {
	warnings := append([]string(nil), report.Warnings...)
	if report.ArtifactPath != "" {
		name := artifactName(version, report.Platform, report.ArtifactPath)
		targetPath := filepath.Join(releaseDir, name)
		if err := copyFile(report.ArtifactPath, targetPath); err != nil {
			return "", nil, err
		}
		return targetPath, warnings, nil
	}

	targetPath := filepath.Join(releaseDir, artifactName(version, report.Platform, "wix-inputs.zip"))
	if err := zipDir(report.OutputDir, targetPath); err != nil {
		return "", nil, err
	}
	warnings = append(warnings, "distribution artifact was not produced; release includes packaged build inputs instead")
	return targetPath, warnings, nil
}

func artifactName(version, platform, sourcePath string) string {
	extension := artifactExtension(sourcePath)
	return "school-gate-installer-v" + version + "-" + platform + "-x64" + extension
}

func artifactExtension(path string) string {
	if strings.HasSuffix(path, ".tar.gz") {
		return ".tar.gz"
	}
	return filepath.Ext(path)
}

func copyFile(sourcePath, targetPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	info, err := source.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}
	target, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer target.Close()
	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return os.Chmod(targetPath, info.Mode())
}

func zipDir(sourceDir, targetPath string) error {
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := zip.NewWriter(file)
	defer writer.Close()

	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil || entry.IsDir() {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		header, err := zip.FileInfoHeader(mustInfo(entry))
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relativePath)
		record, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = record.Write(data)
		return err
	})
}

func mustInfo(entry fs.DirEntry) fs.FileInfo {
	info, err := entry.Info()
	if err != nil {
		panic(err)
	}
	return info
}
