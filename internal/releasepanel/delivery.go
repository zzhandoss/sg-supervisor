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
	"sort"
	"strings"
)

type deliveryFiles struct {
	PayloadPath   string
	ArchivePath   string
	MetadataPath  string
	ChecksumsPath string
}

func buildDeliveryRelease(ctx context.Context, root, version, platform, deliveryRoot string, warnings []string) (LocalReleaseReport, error) {
	version = trimVersion(version)
	releaseDir := filepath.Join(root, "releases", "v"+version, platform)
	if err := os.RemoveAll(releaseDir); err != nil {
		return LocalReleaseReport{}, err
	}
	if err := os.MkdirAll(releaseDir, 0o755); err != nil {
		return LocalReleaseReport{}, err
	}

	files, err := materializeDeliveryFiles(releaseDir, version, platform)
	if err != nil {
		return LocalReleaseReport{}, err
	}
	if err := writeDeliveryReadme(releaseDir); err != nil {
		return LocalReleaseReport{}, err
	}
	if err := writeDeliveryArchive(ctx, files.ArchivePath, deliveryRoot); err != nil {
		return LocalReleaseReport{}, err
	}
	if err := writeDeliveryChecksums(files.ChecksumsPath, files.ArchivePath); err != nil {
		return LocalReleaseReport{}, err
	}
	if err := writeDeliveryMetadata(files.MetadataPath, version, platform, files, warnings); err != nil {
		return LocalReleaseReport{}, err
	}

	return LocalReleaseReport{
		Version:       version,
		Platform:      platform,
		ReleaseDir:    releaseDir,
		ArtifactPath:  files.ArchivePath,
		MetadataPath:  files.MetadataPath,
		ChecksumsPath: files.ChecksumsPath,
		Warnings:      append([]string(nil), warnings...),
	}, nil
}

func buildLocalReleaseSet(root, version string, reports []LocalReleaseReport) (LocalReleaseSetReport, error) {
	version = trimVersion(version)
	releaseDir := filepath.Join(root, "releases", "v"+version)
	sort.Slice(reports, func(i, j int) bool {
		return reports[i].Platform < reports[j].Platform
	})
	platforms := make([]string, 0, len(reports))
	warnings := make([]string, 0)
	for _, report := range reports {
		platforms = append(platforms, report.Platform)
		warnings = append(warnings, report.Warnings...)
	}
	metadataPath, err := writeLocalReleaseSetMetadata(releaseDir, version, reports, warnings)
	if err != nil {
		return LocalReleaseSetReport{}, err
	}
	return LocalReleaseSetReport{
		Version:      version,
		Platforms:    platforms,
		ReleaseDir:   releaseDir,
		MetadataPath: metadataPath,
		Reports:      reports,
		Warnings:     warnings,
	}, nil
}

func writeLocalReleaseSetMetadata(releaseDir, version string, reports []LocalReleaseReport, warnings []string) (string, error) {
	body := map[string]any{
		"version":   version,
		"reports":   reports,
		"warnings":  warnings,
		"platforms": collectReleasePlatforms(reports),
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return "", err
	}
	data = append(data, '\n')
	path := filepath.Join(releaseDir, "release-set.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func collectReleasePlatforms(reports []LocalReleaseReport) []string {
	platforms := make([]string, 0, len(reports))
	for _, report := range reports {
		platforms = append(platforms, report.Platform)
	}
	sort.Strings(platforms)
	return platforms
}

func materializeDeliveryFiles(releaseDir, version, platform string) (deliveryFiles, error) {
	files := deliveryFiles{
		PayloadPath:   filepath.Join(releaseDir, deliveryArtifactName(version, platform)),
		ArchivePath:   filepath.Join(releaseDir, deliveryArtifactName(version, platform)),
		MetadataPath:  filepath.Join(releaseDir, deliverySupportName(version, platform, "release.json")),
		ChecksumsPath: filepath.Join(releaseDir, deliverySupportName(version, platform, "SHA256SUMS.txt")),
	}
	return files, nil
}

func writeDeliveryArchive(ctx context.Context, targetPath, deliveryRoot string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return zipDir(deliveryRoot, targetPath)
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
		"warnings":      warnings,
	}
	data, err := json.MarshalIndent(body, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeDeliveryReadme(releaseDir string) error {
	lines := []string{
		"School Gate delivery package",
		"",
		"1. Extract this archive.",
		"2. Run sg-supervisor from the extracted directory.",
		"3. Open the local Control Center and start bootstrap installation.",
		"4. Wait for dependency install and build steps to finish, then start the application manually.",
	}
	return os.WriteFile(filepath.Join(releaseDir, "README.txt"), []byte(strings.Join(lines, "\n")+"\n"), 0o644)
}

func deliveryArtifactName(version, platform string) string {
	return "school-gate-delivery-v" + version + "-" + platform + "-x64.zip"
}

func deliverySupportName(version, platform, suffix string) string {
	suffix = strings.TrimPrefix(suffix, "-")
	return "school-gate-delivery-v" + version + "-" + platform + "-x64-" + suffix
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
