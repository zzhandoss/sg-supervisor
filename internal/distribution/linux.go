package distribution

import (
	"archive/tar"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"

	"sg-supervisor/internal/packaging"
)

func buildLinux(root string, stage packaging.AssembleReport) (Report, error) {
	outputDir := filepath.Join(root, "dist", "linux")
	packageDir := filepath.Join(outputDir, "package")
	if err := os.RemoveAll(outputDir); err != nil {
		return Report{}, err
	}
	if err := os.MkdirAll(packageDir, 0o755); err != nil {
		return Report{}, err
	}

	payloadDir := filepath.Join(packageDir, "payload")
	if err := copyTree(stage.OutputDir, payloadDir); err != nil {
		return Report{}, err
	}
	installPath := filepath.Join(packageDir, "install.sh")
	uninstallPath := filepath.Join(packageDir, "uninstall.sh")
	if err := os.WriteFile(installPath, []byte(renderLinuxInstallScript()), 0o755); err != nil {
		return Report{}, err
	}
	if err := os.WriteFile(uninstallPath, []byte(renderLinuxUninstallScript()), 0o755); err != nil {
		return Report{}, err
	}

	artifactPath := filepath.Join(outputDir, "school-gate-linux-x64.tar.gz")
	if err := writeTarGz(artifactPath, packageDir); err != nil {
		return Report{}, err
	}

	return Report{
		Platform:       "linux",
		StageDir:       stage.OutputDir,
		OutputDir:      outputDir,
		ArtifactPath:   artifactPath,
		GeneratedFiles: []string{installPath, uninstallPath, artifactPath},
	}, nil
}

func renderLinuxInstallScript() string {
	return strings.Join([]string{
		"#!/usr/bin/env sh",
		"set -eu",
		`ROOT=""`,
		`while [ "$#" -gt 0 ]; do`,
		`  case "$1" in`,
		`    --root) ROOT="$2"; shift 2 ;;`,
		`    *) echo "unknown argument: $1" >&2; exit 1 ;;`,
		`  esac`,
		`done`,
		`if [ -z "$ROOT" ]; then echo "usage: ./install.sh --root /path/to/school-gate" >&2; exit 1; fi`,
		`SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)`,
		`PAYLOAD_DIR="$SCRIPT_DIR/payload"`,
		`mkdir -p "$ROOT/install" "$ROOT/runtime" "$ROOT/supervisor"`,
		`cp -R "$PAYLOAD_DIR/install/." "$ROOT/install/"`,
		`cp -R "$PAYLOAD_DIR/runtime/." "$ROOT/runtime/"`,
		`cp -R "$PAYLOAD_DIR/supervisor/." "$ROOT/supervisor/"`,
		`cp "$ROOT/runtime/service-host/linux/school-gate-supervisor.service" /etc/systemd/system/school-gate-supervisor.service`,
		`systemctl daemon-reload`,
		`systemctl enable --now school-gate-supervisor`,
		"",
	}, "\n")
}

func renderLinuxUninstallScript() string {
	return strings.Join([]string{
		"#!/usr/bin/env sh",
		"set -eu",
		`ROOT=""`,
		`PURGE="false"`,
		`while [ "$#" -gt 0 ]; do`,
		`  case "$1" in`,
		`    --root) ROOT="$2"; shift 2 ;;`,
		`    --purge) PURGE="true"; shift 1 ;;`,
		`    *) echo "unknown argument: $1" >&2; exit 1 ;;`,
		`  esac`,
		`done`,
		`if [ -z "$ROOT" ]; then echo "usage: ./uninstall.sh --root /path/to/school-gate [--purge]" >&2; exit 1; fi`,
		`systemctl disable --now school-gate-supervisor || true`,
		`rm -f /etc/systemd/system/school-gate-supervisor.service`,
		`systemctl daemon-reload`,
		`rm -rf "$ROOT/install" "$ROOT/runtime" "$ROOT/supervisor"`,
		`if [ "$PURGE" = "true" ]; then rm -rf "$ROOT/config" "$ROOT/data" "$ROOT/logs" "$ROOT/licenses" "$ROOT/backups" "$ROOT/updates"; fi`,
		"",
	}, "\n")
}

func writeTarGz(path, sourceDir string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return nil
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relativePath)
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		_, err = tarWriter.Write(data)
		return err
	})
}
