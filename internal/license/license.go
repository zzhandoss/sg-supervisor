package license

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"sg-supervisor/internal/config"
)

type Payload struct {
	LicenseID   string    `json:"licenseId"`
	Customer    string    `json:"customer"`
	Mode        string    `json:"mode"`
	Edition     string    `json:"edition"`
	Features    []string  `json:"features"`
	ExpiresAt   time.Time `json:"expiresAt"`
	IssuedAt    time.Time `json:"issuedAt"`
	Perpetual   bool      `json:"perpetual"`
	Fingerprint string    `json:"fingerprint,omitempty"`
}

type File struct {
	Payload   Payload `json:"payload"`
	Signature string  `json:"signature"`
}

type ActivationRequest struct {
	Product      string    `json:"product"`
	RequestedAt  time.Time `json:"requestedAt"`
	CustomerHint string    `json:"customerHint,omitempty"`
	Fingerprint  string    `json:"fingerprint"`
	Signals      []string  `json:"signals"`
}

type Status struct {
	Valid       bool   `json:"valid"`
	Mode        string `json:"mode,omitempty"`
	Customer    string `json:"customer,omitempty"`
	ExpiresAt   string `json:"expiresAt,omitempty"`
	LicensePath string `json:"licensePath,omitempty"`
	Reason      string `json:"reason,omitempty"`
}

type Store struct {
	layout config.Layout
	cfg    config.SupervisorConfig
}

func NewStore(layout config.Layout, cfg config.SupervisorConfig) *Store {
	return &Store{layout: layout, cfg: cfg}
}

func (s *Store) Status(ctx context.Context) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	path := filepath.Join(s.layout.LicensesDir, "current-license.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return Status{Valid: false, Reason: "license file not found", LicensePath: path}, nil
	} else if err != nil {
		return Status{}, err
	}

	licenseFile, err := Read(path)
	if err != nil {
		return Status{Valid: false, Reason: err.Error(), LicensePath: path}, nil
	}
	if err := Verify(licenseFile, s.cfg.PublicKeyBase64); err != nil {
		return Status{Valid: false, Reason: err.Error(), LicensePath: path}, nil
	}
	if err := validateAgainstHost(licenseFile.Payload); err != nil {
		return Status{Valid: false, Reason: err.Error(), LicensePath: path}, nil
	}
	return Status{
		Valid:       true,
		Mode:        licenseFile.Payload.Mode,
		Customer:    licenseFile.Payload.Customer,
		ExpiresAt:   licenseFile.Payload.ExpiresAt.Format(time.RFC3339),
		LicensePath: path,
	}, nil
}

func (s *Store) Import(ctx context.Context, sourcePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	licenseFile, err := Read(sourcePath)
	if err != nil {
		return err
	}
	if err := Verify(licenseFile, s.cfg.PublicKeyBase64); err != nil {
		return err
	}
	if err := validateAgainstHost(licenseFile.Payload); err != nil {
		return err
	}
	target := filepath.Join(s.layout.LicensesDir, "current-license.json")
	return WriteFile(target, licenseFile)
}

func BuildActivationRequest(customer string) (ActivationRequest, error) {
	fingerprint, signals, err := ComputeFingerprint()
	if err != nil {
		return ActivationRequest{}, err
	}
	return ActivationRequest{
		Product:      "School Gate",
		RequestedAt:  time.Now().UTC(),
		CustomerHint: customer,
		Fingerprint:  fingerprint,
		Signals:      signals,
	}, nil
}

func WriteActivationRequest(path string, request ActivationRequest) error {
	data, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func Read(path string) (File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	var file File
	if err := json.Unmarshal(data, &file); err != nil {
		return File{}, err
	}
	return file, nil
}

func WriteFile(path string, file File) error {
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func Sign(payload Payload, privateKey ed25519.PrivateKey) (File, error) {
	payloadBytes, err := canonicalPayload(payload)
	if err != nil {
		return File{}, err
	}
	signature := ed25519.Sign(privateKey, payloadBytes)
	return File{Payload: payload, Signature: base64.StdEncoding.EncodeToString(signature)}, nil
}

func Verify(file File, publicKeyBase64 string) error {
	if publicKeyBase64 == "" {
		return errors.New("publicKeyBase64 is empty")
	}
	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return err
	}
	signature, err := base64.StdEncoding.DecodeString(file.Signature)
	if err != nil {
		return err
	}
	payloadBytes, err := canonicalPayload(file.Payload)
	if err != nil {
		return err
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKeyBytes), payloadBytes, signature) {
		return errors.New("license signature is invalid")
	}
	if !file.Payload.Perpetual && time.Now().UTC().After(file.Payload.ExpiresAt) {
		return errors.New("license has expired")
	}
	return nil
}

func validateAgainstHost(payload Payload) error {
	if payload.Mode != "bound" {
		return nil
	}
	fingerprint, _, err := ComputeFingerprint()
	if err != nil {
		return err
	}
	if fingerprint != payload.Fingerprint {
		return errors.New("license fingerprint does not match this machine")
	}
	return nil
}

func canonicalPayload(payload Payload) ([]byte, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(data)
	return sum[:], nil
}
