package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteInstalledEnvFilesCreatesCanonicalEnvFiles(t *testing.T) {
	layout := NewLayout(t.TempDir())
	catalog := ServiceCatalog{
		Services: []ServiceSpec{
			{
				Name: "api",
				Env: map[string]string{
					"NODE_ENV":   "production",
					"LOG_DIR":    filepath.Join(layout.LogsDir, "school-gate"),
					"BACKUP_DIR": filepath.Join(layout.BackupsDir, "school-gate"),
				},
			},
			{
				Name: "bot",
				Env: map[string]string{
					"TELEGRAM_BOT_TOKEN": "bot-token",
				},
			},
			{
				Name: "dahua-terminal-adapter",
				Env: map[string]string{
					"LOG_OUTPUT": "file",
					"LOG_DIR":    filepath.Join(layout.LogsDir, "dahua-terminal-adapter"),
				},
			},
		},
	}

	if err := WriteInstalledEnvFiles(layout, catalog); err != nil {
		t.Fatalf("write installed env files: %v", err)
	}

	coreData, err := os.ReadFile(SchoolGateEnvFile(layout))
	if err != nil {
		t.Fatalf("read school-gate env: %v", err)
	}
	coreContent := string(coreData)
	if !strings.Contains(coreContent, "LOG_DIR="+filepath.Join(layout.LogsDir, "school-gate")+"\n") {
		t.Fatalf("missing school-gate log dir in env: %q", coreContent)
	}
	if !strings.Contains(coreContent, "TELEGRAM_BOT_TOKEN=bot-token\n") {
		t.Fatalf("missing telegram token in env: %q", coreContent)
	}

	adapterData, err := os.ReadFile(AdapterEnvFile(layout))
	if err != nil {
		t.Fatalf("read adapter env: %v", err)
	}
	adapterContent := string(adapterData)
	if !strings.Contains(adapterContent, "LOG_OUTPUT=file\n") {
		t.Fatalf("missing adapter log output in env: %q", adapterContent)
	}
}
