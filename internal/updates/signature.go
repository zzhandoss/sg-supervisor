package updates

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"sg-supervisor/internal/config"
)

func verifyManifestSignature(cfg config.SupervisorConfig, manifestData, signatureData []byte) error {
	publicKeyBase64 := cfg.PackageSigningPublicKeyBase64
	if publicKeyBase64 == "" {
		publicKeyBase64 = cfg.PublicKeyBase64
	}
	if publicKeyBase64 == "" {
		return errors.New("package signing public key is not configured")
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	if err != nil {
		return err
	}
	signature, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(signatureData)))
	if err != nil {
		return err
	}
	if !ed25519.Verify(ed25519.PublicKey(publicKey), manifestData, signature) {
		return errors.New("package signature is invalid")
	}
	return nil
}

func readSignatureFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
