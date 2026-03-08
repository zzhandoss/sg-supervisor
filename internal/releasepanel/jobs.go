package releasepanel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type JobStore struct {
	layout Layout
}

func NewJobStore(layout Layout) *JobStore {
	return &JobStore{layout: layout}
}

func (s *JobStore) Save(job Job) error {
	data, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(s.layout.JobsDir, job.ID+".json"), data, 0o644)
}

func (s *JobStore) List() ([]Job, error) {
	entries, err := os.ReadDir(s.layout.JobsDir)
	if err != nil {
		return nil, err
	}
	jobs := make([]Job, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(s.layout.JobsDir, entry.Name()))
		if err != nil {
			return nil, err
		}
		var job Job
		if err := json.Unmarshal(data, &job); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt > jobs[j].CreatedAt
	})
	return jobs, nil
}

func (s *JobStore) RecoverInterrupted() error {
	jobs, err := s.List()
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if job.Status != JobStatusQueued && job.Status != JobStatusRunning {
			continue
		}
		job.Status = JobStatusFailed
		job.Error = "job was interrupted"
		job.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		job.Logs = append(job.Logs, "job was interrupted")
		if err := s.Save(job); err != nil {
			return err
		}
	}
	return nil
}
