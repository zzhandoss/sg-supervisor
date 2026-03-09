package releasepanel

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sg-supervisor/internal/release"
)

type deliveryFiles struct {
	InstallerPath string
	PayloadPath   string
	ArchivePath   string
	MetadataPath  string
	ChecksumsPath string
}

func buildDeliveryRelease(ctx context.Context, root, version, platform, installerPath, payloadPath string, warnings []string) (release.Report, error) {
	version = trimVersion(version)
	releaseDir := filepath.Join(root, "releases", "v"+version, platform)
	if err := os.RemoveAll(releaseDir); err != nil {
		return release.Report{}, err
	}
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		return release.Report{}, err
	}

	files, err := materializeDeliveryFiles(ctx, releaseDir, version, platform, installerPath, payloadPath)
	if err != nil {
		return release.Report{}, err
	}
	if err := writeDeliveryReadme(releaseDir, filepath.Base(files.InstallerPath), filepath.Base(files.PayloadPath)); err != nil {
		return release.Report{}, err
	}
	if err := writeDeliveryArchive(ctx, files.ArchivePath, files.InstallerPath, files.PayloadPath, filepath.Join(releaseDir, "README.txt")); err != nil {
		return release.Report{}, err
	}
	if err := writeDeliveryChecksums(files.ChecksumsPath, files.ArchivePath, files.InstallerPath, files.PayloadPath); err != nil {
		return release.Report{}, err
	}
	if err := writeDeliveryMetadata(files.MetadataPath, version, platform, files, warnings); err != nil {
		return release.Report{}, err
	}

	return release.Report{
		Version:       version,
		Platform:      platform,
		ReleaseDir:    releaseDir,
		ArtifactPath:  files.ArchivePath,
		MetadataPath:  files.MetadataPath,
		ChecksumsPath: files.ChecksumsPath,
		Warnings:      append([]string(nil), warnings...),
	}, nil
}

func materializeDeliveryFiles(ctx context.Context, releaseDir, version, platform, installerPath, payloadPath string) (deliveryFiles, error) {
	files := deliveryFiles{
		InstallerPath: filepath.Join(releaseDir, installerArtifactName(version, platform, installerPath)),
		PayloadPath:   filepath.Join(releaseDir, payloadArtifactName(version, platform)),
		ArchivePath:   filepath.Join(releaseDir, deliveryArtifactName(version, platform)),
		MetadataPath:  filepath.Join(releaseDir, deliverySupportName(version, platform, "release.json")),
		ChecksumsPath: filepath.Join(releaseDir, deliverySupportName(version, platform, "SHA256SUMS.txt")),
	}
	if err := ctx.Err(); err != nil {
		return deliveryFiles{}, err
	}
	if err := copyArtifact(installerPath, files.InstallerPath); err != nil {
		return deliveryFiles{}, err
	}
	if err := copyArtifact(payloadPath, files.PayloadPath); err != nil {
		return deliveryFiles{}, err
	}
	return files, nil
}

func writeDeliveryArchive(ctx context.Context, targetPath, installerPath, payloadPath, readmePath string) error {
	archive, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer archive.Close()
	writer := zip.NewWriter(archive)
	defer writer.Close()
	entries := []struct {
		sourcePath string
		targetPath string
	}{
		{sourcePath: installerPath, targetPath: filepath.Base(installerPath)},
		{sourcePath: payloadPath, targetPath: filepath.ToSlash(filepath.Join("payload", filepath.Base(payloadPath)))},
		{sourcePath: readmePath, targetPath: "README.txt"},
	}
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		info, err := os.Stat(entry.sourcePath)
		if err != nil {
			return err
		}
		if err := writeZipFile(writer, entry.sourcePath, entry.targetPath, info.Mode()); err != nil {
			return err
		}
	}
	return nil
}

func writeDeliveryChecksums(path string, files ...string) error {
	var builder strings.Builder
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		builder.WriteString(hex.EncodeToString(sum[:]))
		builder.WriteString("  ")
		builder.WriteString(filepath.Base(file))
		builder.WriteByte('\n')
	}
	return os.WriteFile(path, []byte(builder.String()), 0o644)
}

func writeDeliveryMetadata(path, version, platform string, files deliveryFiles, warnings []string) error {
	body := map[string]any{
		"version":       version,
		"platform":      platform,
		"artifactPath":  files.ArchivePath,
		"checksumsPath": files.ChecksumsPath,
		"installerPath": files.InstallerPath,
		"payloadPath":   files.PayloadPath,
		"warnings":      warnings,
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeDeliveryReadme(releaseDir, installerName, payloadName string) error {
	lines := []string{
		"School Gate delivery package",
		"",
		"1. Extract this archive.",
		"2. Run the installer.",
		"3. Open the local Control Center after installation.",
		"4. Import and apply the local payload bundle from the extracted payload directory.",
		"",
		"Installer: " + installerName,
		"Payload: payload/" + payloadName,
	}
	return os.WriteFile(filepath.Join(releaseDir, "README.txt"), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func installerArtifactName(version, platform, sourcePath string) string {
	return "school-gate-installer-v" + version + "-" + platform + "-x64" + artifactExtension(sourcePath)
}

func payloadArtifactName(version, platform string) string {
	return "school-gate-package-v" + version + "-" + platform + "-x64.zip"
}

func deliveryArtifactName(version, platform string) string {
	return "school-gate-delivery-v" + version + "-" + platform + "-x64.zip"
}

func deliverySupportName(version, platform, suffix string) string {
	suffix = strings.TrimPrefix(suffix, "-")
	return "school-gate-delivery-v" + version + "-" + platform + "-x64-" + suffix
}

func artifactExtension(path string) string {
	if strings.HasSuffix(path, ".tar.gz") {
		return ".tar.gz"
	}
	return filepath.Ext(path)
}

func copyArtifact(sourcePath, targetPath string) error {
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
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return writeZipFile(writer, path, filepath.ToSlash(relativePath), info.Mode())
	})
}
