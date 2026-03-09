package app

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func extractBootstrapArchive(sourcePath, targetDir string) error {
	if strings.HasSuffix(sourcePath, ".zip") {
		return extractBootstrapZip(sourcePath, targetDir)
	}
	if strings.HasSuffix(sourcePath, ".tar.gz") {
		return extractBootstrapTarGz(sourcePath, targetDir)
	}
	if strings.HasSuffix(sourcePath, ".tar.xz") {
		return extractBootstrapTarXz(sourcePath, targetDir)
	}
	return errors.New("unsupported archive type: " + sourcePath)
}

func extractBootstrapZip(sourcePath, targetDir string) error {
	reader, err := zip.OpenReader(sourcePath)
	if err != nil {
		return err
	}
	defer reader.Close()
	prefix := commonZipPrefix(reader.File)
	for _, entry := range reader.File {
		name := trimArchivePrefix(entry.Name, prefix)
		if name == "" {
			continue
		}
		targetPath := filepath.Join(targetDir, filepath.FromSlash(name))
		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		source, err := entry.Open()
		if err != nil {
			return err
		}
		if err := writeBootstrapReader(targetPath, source, entry.Mode()); err != nil {
			source.Close()
			return err
		}
		source.Close()
	}
	return nil
}

func extractBootstrapTarGz(sourcePath, targetDir string) error {
	file, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer file.Close()
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzipReader.Close()
	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		targetPath := filepath.Join(targetDir, filepath.FromSlash(header.Name))
		if header.FileInfo().IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if err := writeBootstrapReader(targetPath, tarReader, header.FileInfo().Mode()); err != nil {
			return err
		}
	}
}

func extractBootstrapTarXz(sourcePath, targetDir string) error {
	command := exec.Command("tar", "-xf", sourcePath, "-C", targetDir)
	output, err := command.CombinedOutput()
	if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		return errors.New("tar extraction failed: " + message)
	}
	return nil
}

func writeBootstrapReader(path string, source io.Reader, mode fs.FileMode) error {
	target, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer target.Close()
	_, err = io.Copy(target, source)
	return err
}

func commonZipPrefix(files []*zip.File) string {
	if len(files) == 0 {
		return ""
	}
	prefix := firstArchiveSegment(files[0].Name)
	if prefix == "" {
		return ""
	}
	for _, file := range files[1:] {
		if firstArchiveSegment(file.Name) != prefix {
			return ""
		}
	}
	return prefix
}

func firstArchiveSegment(path string) string {
	path = strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
	if path == "" {
		return ""
	}
	parts := strings.Split(path, "/")
	return parts[0]
}

func trimArchivePrefix(path, prefix string) string {
	path = strings.Trim(strings.ReplaceAll(path, "\\", "/"), "/")
	if prefix == "" {
		return path
	}
	path = strings.TrimPrefix(path, prefix)
	return strings.Trim(path, "/")
}
