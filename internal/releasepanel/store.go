package releasepanel

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"os"
)

type Store struct {
	layout Layout
}

func NewStore(layout Layout) *Store {
	return &Store{layout: layout}
}

func (s *Store) Ensure(repoRoot string) (State, error) {
	if err := EnsureLayout(s.layout); err != nil {
		return State{}, err
	}
	if _, err := os.Stat(s.layout.StatePath); os.IsNotExist(err) {
		state, err := defaultState(repoRoot)
		if err != nil {
			return State{}, err
		}
		return state, s.Save(state)
	} else if err != nil {
		return State{}, err
	}
	state, err := s.Load()
	if err != nil {
		return State{}, err
	}
	if state.RepoRoot == "" && repoRoot != "" {
		state.RepoRoot = repoRoot
	}
	if err := ensureKeys(&state); err != nil {
		return State{}, err
	}
	return state, s.Save(state)
}

func (s *Store) Load() (State, error) {
	data, err := os.ReadFile(s.layout.StatePath)
	if err != nil {
		return State{}, err
	}
	var state State
	if err := json.Unmarshal(data, &state); err != nil {
		return State{}, err
	}
	return state, nil
}

func (s *Store) Save(state State) error {
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.layout.StatePath, data, 0o644)
}

func defaultState(repoRoot string) (State, error) {
	state := State{
		ListenAddress: "127.0.0.1:8790",
		RepoRoot:      repoRoot,
	}
	return state, ensureKeys(&state)
}

func ensureKeys(state *State) error {
	if state.Keys.LicensePrivateKeyBase64 == "" || state.Keys.LicensePublicKeyBase64 == "" {
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		state.Keys.LicensePrivateKeyBase64 = base64.StdEncoding.EncodeToString(privateKey)
		state.Keys.LicensePublicKeyBase64 = base64.StdEncoding.EncodeToString(publicKey)
	}
	if state.Keys.PackagePrivateKeyBase64 == "" || state.Keys.PackagePublicKeyBase64 == "" {
		publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			return err
		}
		state.Keys.PackagePrivateKeyBase64 = base64.StdEncoding.EncodeToString(privateKey)
		state.Keys.PackagePublicKeyBase64 = base64.StdEncoding.EncodeToString(publicKey)
	}
	return nil
}
