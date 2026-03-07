package updates

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"
)

type ActiveRecord struct {
	PackageID         string   `json:"packageId"`
	AppliedAt         string   `json:"appliedAt"`
	ProductVersion    string   `json:"productVersion"`
	CoreVersion       string   `json:"coreVersion"`
	SupervisorVersion string   `json:"supervisorVersion"`
	Adapters          []string `json:"adapters"`
	ManifestPath      string   `json:"manifestPath"`
	BackupPath        string   `json:"backupPath,omitempty"`
	StoppedServices   []string `json:"stoppedServices,omitempty"`
	RestartedServices []string `json:"restartedServices,omitempty"`
}

func (s *Store) Apply(ctx context.Context, packageID string) (ActiveRecord, error) {
	if err := ctx.Err(); err != nil {
		return ActiveRecord{}, err
	}
	records, err := s.List(ctx)
	if err != nil {
		return ActiveRecord{}, err
	}

	for _, record := range records {
		if record.PackageID != packageID {
			continue
		}
		if record.SourceType == "bundle" {
			if err := s.applyBundle(ctx, record); err != nil {
				return ActiveRecord{}, err
			}
		}
		active := ActiveRecord{
			PackageID:         record.PackageID,
			AppliedAt:         time.Now().UTC().Format(time.RFC3339),
			ProductVersion:    record.Manifest.ProductVersion,
			CoreVersion:       record.Manifest.CoreVersion,
			SupervisorVersion: record.Manifest.SupervisorVersion,
			Adapters:          adapterLabels(record),
			ManifestPath:      record.StoredPath,
		}
		if record.SourceType == "bundle" {
			active.BackupPath = filepath.Join(s.layout.BackupsDir, record.PackageID)
		}
		if err := s.writeActive(active); err != nil {
			return ActiveRecord{}, err
		}
		return active, nil
	}

	return ActiveRecord{}, errors.New("package id not found")
}

func (s *Store) Active(ctx context.Context) (ActiveRecord, error) {
	if err := ctx.Err(); err != nil {
		return ActiveRecord{}, err
	}
	data, err := os.ReadFile(s.activePath())
	if os.IsNotExist(err) {
		return ActiveRecord{}, nil
	}
	if err != nil {
		return ActiveRecord{}, err
	}

	var active ActiveRecord
	if err := json.Unmarshal(data, &active); err != nil {
		return ActiveRecord{}, err
	}
	return active, nil
}

func (s *Store) activePath() string {
	return filepath.Join(s.layout.UpdatesDir, "active-package.json")
}

func (s *Store) writeActive(active ActiveRecord) error {
	data, err := json.MarshalIndent(active, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.activePath(), data, 0o644)
}

func (s *Store) SaveActive(active ActiveRecord) error {
	return s.writeActive(active)
}

func (s *Store) ClearActive() error {
	if err := os.Remove(s.activePath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func adapterLabels(record Record) []string {
	adapters := make([]string, 0, len(record.Manifest.Adapters))
	for _, adapter := range record.Manifest.Adapters {
		adapters = append(adapters, adapter.Key+"@"+adapter.Version)
	}
	return adapters
}
