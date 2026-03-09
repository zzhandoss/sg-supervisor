package releasepanel

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

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

func writeZipFile(writer *zip.Writer, sourcePath, targetPath string, mode fs.FileMode) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer source.Close()
	header := &zip.FileHeader{
		Name:   filepath.ToSlash(targetPath),
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
