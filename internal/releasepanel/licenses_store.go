package releasepanel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (s *Store) SaveIssuedLicense(record IssuedLicenseRecord) error {
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	path := filepath.Join(s.layout.LicensesDir, "issued", record.LicenseID+".record.json")
	return os.WriteFile(path, data, 0o644)
}

func (s *Store) ListIssuedLicenses() ([]IssuedLicenseRecord, error) {
	dir := filepath.Join(s.layout.LicensesDir, "issued")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	records := make([]IssuedLicenseRecord, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".record.json") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var record IssuedLicenseRecord
		if err := json.Unmarshal(data, &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].IssuedAt > records[j].IssuedAt
	})
	return records, nil
}
