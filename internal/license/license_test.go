package license

import (
	"crypto/ed25519"
	"encoding/base64"
	"path/filepath"
	"testing"
	"time"

	"sg-supervisor/internal/config"
)

func TestLicenseRoundTrip(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate keypair: %v", err)
	}

	payload := Payload{
		LicenseID: "lic-1",
		Customer:  "ACME",
		Mode:      "free",
		Edition:   "standard",
		Features:  []string{"core"},
		ExpiresAt: time.Now().UTC().Add(24 * time.Hour),
		IssuedAt:  time.Now().UTC(),
	}
	file, err := Sign(payload, privateKey)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	root := t.TempDir()
	layout := config.NewLayout(root)
	if err := config.EnsureLayout(layout); err != nil {
		t.Fatalf("ensure layout: %v", err)
	}
	store := NewStore(layout, config.SupervisorConfig{
		ProductName:     "School Gate",
		PublicKeyBase64: base64.StdEncoding.EncodeToString(publicKey),
	})

	sourcePath := filepath.Join(root, "license.json")
	if err := WriteFile(sourcePath, file); err != nil {
		t.Fatalf("write license: %v", err)
	}
	if err := store.Import(t.Context(), sourcePath); err != nil {
		t.Fatalf("import: %v", err)
	}

	status, err := store.Status(t.Context())
	if err != nil {
		t.Fatalf("status: %v", err)
	}
	if !status.Valid {
		t.Fatalf("expected valid license, got %+v", status)
	}
}
