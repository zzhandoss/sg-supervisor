package releasepanel

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NodeDistSource struct {
	baseURL string
	client  *http.Client
}

type NodeSource interface {
	ListVersions(ctx context.Context) ([]ReleaseVersion, error)
	Download(version, platform, targetDir string) (string, error)
}

func NewNodeDistSource() *NodeDistSource {
	return &NodeDistSource{
		baseURL: "https://nodejs.org/dist",
		client: &http.Client{
			Timeout: 2 * time.Minute,
		},
	}
}

func (s *NodeDistSource) ListVersions(ctx context.Context) ([]ReleaseVersion, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/index.json", nil)
	if err != nil {
		return nil, err
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, errors.New("node version index request failed with status " + response.Status)
	}
	var payload []struct {
		Version string `json:"version"`
		Date    string `json:"date"`
		LTS     any    `json:"lts"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, err
	}
	versions := make([]ReleaseVersion, 0, min(20, len(payload)))
	for index, item := range payload {
		if index == 20 {
			break
		}
		name := ""
		switch value := item.LTS.(type) {
		case string:
			name = value
		case bool:
			if value {
				name = "LTS"
			}
		}
		versions = append(versions, ReleaseVersion{
			Tag:         item.Version,
			Name:        name,
			PublishedAt: item.Date,
		})
	}
	return versions, nil
}

func (s *NodeDistSource) Download(version, platform, targetDir string) (string, error) {
	version = normalizeTag(version)
	if version == "" {
		return "", errors.New("node version is required")
	}
	fileName, err := nodeArchiveName(version, platform)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}
	targetPath := filepath.Join(targetDir, fileName)
	url := strings.TrimRight(s.baseURL, "/") + "/" + version + "/" + fileName
	if err := downloadURL(s.client, url, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}

func nodeArchiveName(version, platform string) (string, error) {
	base := "node-" + version
	switch platform {
	case "windows":
		return base + "-win-x64.zip", nil
	case "linux":
		return base + "-linux-x64.tar.xz", nil
	default:
		return "", errors.New("unsupported node platform: " + platform)
	}
}

func downloadURL(client *http.Client, url, targetPath string) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("download failed for " + url + " with status " + response.Status)
	}
	file, err := os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.ReadFrom(response.Body)
	return err
}
