package releasepanel

import (
	"archive/zip"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"sg-supervisor/internal/config"
	"sg-supervisor/internal/manifest"
)

func buildLocalPayloadBundle(ctx context.Context, workspaceRoot string, state State, targetPath string) (manifest.File, error) {
	layout := config.NewLayout(workspaceRoot)
	file := manifest.File{
		ProductVersion:    trimVersion(state.Recipe.InstallerVersion),
		CoreVersion:       trimVersion(state.Recipe.SchoolGateVersion),
		SupervisorVersion: trimVersion(state.Recipe.InstallerVersion),
		Runtime: manifest.Runtime{
			NodeVersion: trimVersion(state.Recipe.NodeVersion),
		},
		Adapters: []manifest.AdapterBundle{
			{
				Key:      "dahua-terminal-adapter",
				Version:  trimVersion(state.Recipe.AdapterVersion),
				Required: true,
			},
		},
		Compatibility: manifest.Compatibility{
			CoreAPI:    1,
			AdapterAPI: 1,
		},
	}
	if err := manifest.Validate(file); err != nil {
		return manifest.File{}, err
	}

	manifestData, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return manifest.File{}, err
	}
	manifestData = append(manifestData, '\n')
	signatureData, err := signPackageManifest(state.Keys.PackagePrivateKeyBase64, manifestData)
	if err != nil {
		return manifest.File{}, err
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return manifest.File{}, err
	}
	archive, err := os.Create(targetPath)
	if err != nil {
		return manifest.File{}, err
	}
	defer archive.Close()

	writer := zip.NewWriter(archive)
	defer writer.Close()

	if err := writeZipBytes(writer, "manifest.json", manifestData, 0o644); err != nil {
		return manifest.File{}, err
	}
	if err := writeZipBytes(writer, "manifest.sig", signatureData, 0o644); err != nil {
		return manifest.File{}, err
	}

	if err := addTreeToZip(ctx, writer, filepath.Join(layout.InstallDir, "core"), "payload/core"); err != nil {
		return manifest.File{}, err
	}
	if err := addTreeToZip(ctx, writer, filepath.Join(layout.InstallDir, "adapters", "dahua-terminal-adapter"), "payload/adapters/dahua-terminal-adapter"); err != nil {
		return manifest.File{}, err
	}
	return file, nil
}

func signPackageManifest(value string, manifestData []byte) ([]byte, error) {
	privateKey, err := decodeSigningPrivateKey(value)
	if err != nil {
		return nil, err
	}
	signature := ed25519.Sign(privateKey, manifestData)
	encoded := base64.StdEncoding.EncodeToString(signature)
	return []byte(encoded + "\n"), nil
}

func decodeSigningPrivateKey(value string) (ed25519.PrivateKey, error) {
	data, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, err
	}
	return ed25519.PrivateKey(data), nil
}

func addTreeToZip(ctx context.Context, writer *zip.Writer, sourceDir, targetRoot string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	info, err := os.Stat(sourceDir)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return writeZipFile(writer, sourceDir, filepath.ToSlash(targetRoot), info.Mode())
	}
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		return writeZipFile(writer, path, filepath.ToSlash(filepath.Join(targetRoot, relativePath)), info.Mode())
	})
}

func writeZipFile(writer *zip.Writer, sourcePath, targetPath string, mode fs.FileMode) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	header := &zip.FileHeader{
		Name:   targetPath,
		Method: zip.Deflate,
	}
	header.SetMode(mode)
	record, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(record, source)
	return err
}

func writeZipBytes(writer *zip.Writer, targetPath string, data []byte, mode fs.FileMode) error {
	header := &zip.FileHeader{
		Name:   filepath.ToSlash(targetPath),
		Method: zip.Deflate,
	}
	header.SetMode(mode)
	record, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = record.Write(data)
	return err
}
