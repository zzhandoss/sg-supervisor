package bootstrap

import (
	"encoding/json"
	"os"
	"path/filepath"

	"sg-supervisor/internal/config"
)

type Store struct {
	dir  string
	path string
}

func NewStore(layout config.Layout) *Store {
	dir := filepath.Join(layout.RuntimeDir, "bootstrap")
	return &Store{
		dir:  dir,
		path: filepath.Join(dir, "status.json"),
	}
}

func (s *Store) Ensure() error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		return s.Save(Status{State: "idle"})
	} else if err != nil {
		return err
	}
	return nil
}

func (s *Store) Load() (Status, error) {
	if err := s.Ensure(); err != nil {
		return Status{}, err
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return Status{}, err
	}
	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return Status{}, err
	}
	if status.State == "" {
		status.State = "idle"
	}
	return status, nil
}

func (s *Store) Save(status Status) error {
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.path, data, 0o644)
}

func (s *Store) Dir() string {
	return s.dir
}
