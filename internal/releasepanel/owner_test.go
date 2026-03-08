package releasepanel

import (
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestOwnerAcquireRejectsActiveLease(t *testing.T) {
	layout := NewLayout(t.TempDir())
	if err := EnsureLayout(layout); err != nil {
		t.Fatal(err)
	}
	store := NewOwnerStore(layout)
	jobs := NewJobStore(layout)
	handle, err := store.Acquire(jobs, "serve")
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()

	_, err = store.Acquire(jobs, "status")
	if !errors.Is(err, errOwnerActive) {
		t.Fatalf("Acquire() error = %v, want %v", err, errOwnerActive)
	}
}

func TestOwnerAcquireRecoversInterruptedJobsOnlyWhenLeaseIsStale(t *testing.T) {
	layout := NewLayout(t.TempDir())
	if err := EnsureLayout(layout); err != nil {
		t.Fatal(err)
	}
	jobs := NewJobStore(layout)
	job := Job{
		ID:        "job-1",
		Type:      JobTypeLocalRelease,
		Status:    JobStatusRunning,
		CreatedAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		StartedAt: time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		Logs:      []string{"local release started"},
	}
	if err := jobs.Save(job); err != nil {
		t.Fatal(err)
	}
	owner := NewOwnerStore(layout)
	stale := OwnerLease{
		ID:          "old-owner",
		PID:         1,
		Purpose:     "serve",
		AcquiredAt:  time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
		HeartbeatAt: time.Now().Add(-(ownerLeaseTTL + time.Second)).UTC().Format(time.RFC3339),
	}
	if err := owner.Save(stale); err != nil {
		t.Fatal(err)
	}

	handle, err := owner.Acquire(jobs, "serve")
	if err != nil {
		t.Fatal(err)
	}
	defer handle.Release()

	recoveredJobs, err := jobs.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveredJobs) != 1 {
		t.Fatalf("jobs count = %d, want 1", len(recoveredJobs))
	}
	if recoveredJobs[0].Status != JobStatusFailed {
		t.Fatalf("job status = %q, want %q", recoveredJobs[0].Status, JobStatusFailed)
	}
	if recoveredJobs[0].Error != "job was interrupted" {
		t.Fatalf("job error = %q", recoveredJobs[0].Error)
	}
	if filepath.Base(layout.OwnerPath) == "" {
		t.Fatal("owner path should be set")
	}
}
