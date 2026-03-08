package releasepanel

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"sg-supervisor/internal/license"
)

func (s *Service) IssueLicense(ctx context.Context, request LicenseIssueRequest) (IssuedLicenseRecord, error) {
	state, err := s.store.Load()
	if err != nil {
		return IssuedLicenseRecord{}, err
	}
	if request.ActivationRequestPath != "" {
		if err := enrichRequestFromActivationRequest(&request, request.ActivationRequestPath); err != nil {
			return IssuedLicenseRecord{}, err
		}
	}
	payload, err := buildLicensePayload(request)
	if err != nil {
		return IssuedLicenseRecord{}, err
	}
	privateKey, err := decodePrivateKey(state.Keys.LicensePrivateKeyBase64)
	if err != nil {
		return IssuedLicenseRecord{}, err
	}
	file, err := license.Sign(payload, privateKey)
	if err != nil {
		return IssuedLicenseRecord{}, err
	}
	targetPath := filepath.Join(s.layout.LicensesDir, "issued", payload.LicenseID+".license.json")
	if err := license.WriteFile(targetPath, file); err != nil {
		return IssuedLicenseRecord{}, err
	}
	record := IssuedLicenseRecord{
		LicenseID:   payload.LicenseID,
		Path:        targetPath,
		Customer:    payload.Customer,
		Mode:        payload.Mode,
		Edition:     payload.Edition,
		Features:    payload.Features,
		ExpiresAt:   payload.ExpiresAt.Format(time.RFC3339),
		IssuedAt:    payload.IssuedAt.Format(time.RFC3339),
		Fingerprint: payload.Fingerprint,
	}
	if payload.Perpetual {
		record.ExpiresAt = ""
	}
	return record, s.store.SaveIssuedLicense(record)
}

func buildLicensePayload(request LicenseIssueRequest) (license.Payload, error) {
	mode := strings.TrimSpace(request.Mode)
	if mode != "free" && mode != "bound" {
		return license.Payload{}, errors.New("mode must be free or bound")
	}
	issuedAt := time.Now().UTC()
	payload := license.Payload{
		LicenseID: newJobID(),
		Customer:  strings.TrimSpace(request.Customer),
		Mode:      mode,
		Edition:   strings.TrimSpace(request.Edition),
		Features:  append([]string(nil), request.Features...),
		IssuedAt:  issuedAt,
		Perpetual: request.Perpetual,
	}
	if payload.Edition == "" {
		payload.Edition = "standard"
	}
	if !request.Perpetual {
		if strings.TrimSpace(request.ExpiresAt) == "" {
			return license.Payload{}, errors.New("expiresAt is required unless perpetual is true")
		}
		expiresAt, err := time.Parse(time.RFC3339, request.ExpiresAt)
		if err != nil {
			return license.Payload{}, err
		}
		payload.ExpiresAt = expiresAt.UTC()
	}
	if mode == "bound" {
		payload.Fingerprint = strings.TrimSpace(request.Fingerprint)
		if payload.Fingerprint == "" {
			return license.Payload{}, errors.New("fingerprint is required for bound licenses")
		}
	}
	return payload, nil
}

func enrichRequestFromActivationRequest(request *LicenseIssueRequest, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var activationRequest license.ActivationRequest
	if err := json.Unmarshal(data, &activationRequest); err != nil {
		return err
	}
	if request.Customer == "" {
		request.Customer = activationRequest.CustomerHint
	}
	if request.Mode == "bound" && request.Fingerprint == "" {
		request.Fingerprint = activationRequest.Fingerprint
	}
	return nil
}

func decodePrivateKey(value string) (ed25519.PrivateKey, error) {
	if value == "" {
		return nil, errors.New("license private key is not configured")
	}
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	return ed25519.PrivateKey(data), nil
}
