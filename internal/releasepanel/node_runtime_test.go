package releasepanel

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareNodeRuntimeWindows(t *testing.T) {
	root := t.TempDir()
	for _, path := range []string{
		filepath.Join(root, "node.exe"),
		filepath.Join(root, "LICENSE"),
		filepath.Join(root, "node_modules", "npm", "package.json"),
		filepath.Join(root, "node_modules", "corepack", "dist", "corepack.js"),
		filepath.Join(root, "corepack.cmd"),
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
		if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
			t.Fatalf("write: %v", err)
		}
	}

	if err := prepareNodeRuntime(root, "windows"); err != nil {
		t.Fatalf("prepare node runtime: %v", err)
	}

	for _, path := range []string{
		filepath.Join(root, "node.exe"),
		filepath.Join(root, "LICENSE"),
		filepath.Join(root, "corepack.cmd"),
		filepath.Join(root, "node_modules", "corepack", "dist", "corepack.js"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected file to remain: %s (%v)", path, err)
		}
	}
	for _, path := range []string{
		filepath.Join(root, "node_modules", "npm"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected path to be removed: %s", path)
		}
	}
}

func TestPrepareNodeRuntimeLinuxNoop(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "bin", "node")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("x"), 0o755); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := prepareNodeRuntime(root, "linux"); err != nil {
		t.Fatalf("prepare node runtime: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected linux runtime to stay untouched: %v", err)
	}
}
