package releasepanel

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func extractArchive(sourcePath, targetDir string) error {
	if strings.HasSuffix(sourcePath, ".zip") {
		return extractZip(sourcePath, targetDir)
	}
	if strings.HasSuffix(sourcePath, ".tar.gz") {
		return extractTarGz(sourcePath, targetDir)
	}
	return errors.New("unsupported archive type: " + sourcePath)
}

func extractZip(sourcePath, targetDir string) error {
	reader, err := zip.OpenReader(sourcePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	prefix := commonArchivePrefixZip(reader.File)
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
		if err := writeReader(targetPath, source, entry.Mode()); err != nil {
			source.Close()
			return err
		}
		source.Close()
	}
	return nil
}

func extractTarGz(sourcePath, targetDir string) error {
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
	entries, err := readTarEntries(tarReader)
	if err != nil {
		return err
	}
	prefix := commonArchivePrefixTar(entries)
	for _, entry := range entries {
		name := trimArchivePrefix(entry.name, prefix)
		if name == "" {
			continue
		}
		targetPath := filepath.Join(targetDir, filepath.FromSlash(name))
		if entry.mode.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(targetPath, entry.data, entry.mode); err != nil {
			return err
		}
	}
	return nil
}

func copyDir(sourceDir, targetDir string) error {
	return filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}
		if relativePath == "." {
			return os.MkdirAll(targetDir, 0o755)
		}
		targetPath := filepath.Join(targetDir, relativePath)
		if entry.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func writeReader(path string, source io.Reader, mode fs.FileMode) error {
	target, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer target.Close()
	_, err = io.Copy(target, source)
	return err
}

type tarEntry struct {
	name string
	mode fs.FileMode
	data []byte
}

func readTarEntries(reader *tar.Reader) ([]tarEntry, error) {
	entries := make([]tarEntry, 0, 32)
	for {
		header, err := reader.Next()
		if errors.Is(err, io.EOF) {
			return entries, nil
		}
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		entries = append(entries, tarEntry{name: header.Name, mode: header.FileInfo().Mode(), data: data})
	}
}

func commonArchivePrefixZip(files []*zip.File) string {
	paths := make([]string, 0, len(files))
	for _, entry := range files {
		paths = append(paths, entry.Name)
	}
	return commonArchivePrefix(paths)
}

func commonArchivePrefixTar(entries []tarEntry) string {
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		paths = append(paths, entry.name)
	}
	return commonArchivePrefix(paths)
}

func commonArchivePrefix(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	prefix := firstArchiveSegment(paths[0])
	if prefix == "" {
		return ""
	}
	for _, path := range paths[1:] {
		if firstArchiveSegment(path) != prefix {
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
