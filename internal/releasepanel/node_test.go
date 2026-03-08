package releasepanel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNodeDistSourceListVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/index.json" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		payload := []map[string]any{
			{"version": "v20.20.1", "date": "2026-03-05", "lts": "Iron"},
			{"version": "v25.8.0", "date": "2026-03-03", "lts": false},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	source := &NodeDistSource{baseURL: server.URL, client: server.Client()}
	versions, err := source.ListVersions(context.Background())
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 2 || versions[0].Tag != "v20.20.1" {
		t.Fatalf("unexpected versions: %+v", versions)
	}
}

func TestNodeDistSourceDownload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v20.20.1/node-v20.20.1-win-x64.zip" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write([]byte("node"))
	}))
	defer server.Close()

	source := &NodeDistSource{baseURL: server.URL, client: server.Client()}
	targetDir := t.TempDir()
	path, err := source.Download("20.20.1", "windows", targetDir)
	if err != nil {
		t.Fatalf("download node: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read node archive: %v", err)
	}
	if string(data) != "node" {
		t.Fatalf("unexpected archive content: %q", string(data))
	}
	if filepath.Dir(path) != targetDir {
		t.Fatalf("unexpected target path: %s", path)
	}
}
