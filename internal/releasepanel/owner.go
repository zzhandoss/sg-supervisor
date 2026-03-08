package releasepanel

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	ownerLeaseTTL      = 15 * time.Second
	ownerHeartbeatStep = 5 * time.Second
)

var errOwnerActive = errors.New("release panel owner is already active")

type OwnerLease struct {
	ID          string `json:"id"`
	PID         int    `json:"pid"`
	Purpose     string `json:"purpose"`
	AcquiredAt  string `json:"acquiredAt"`
	HeartbeatAt string `json:"heartbeatAt"`
}

type OwnerStore struct {
	layout Layout
}

type OwnerHandle struct {
	store  *OwnerStore
	lease  OwnerLease
	stopCh chan struct{}
	doneCh chan struct{}
	once   sync.Once
}

func NewOwnerStore(layout Layout) *OwnerStore {
	return &OwnerStore{layout: layout}
}

func (s *OwnerStore) Load() (OwnerLease, error) {
	data, err := os.ReadFile(s.layout.OwnerPath)
	if err != nil {
		return OwnerLease{}, err
	}
	var lease OwnerLease
	if err := json.Unmarshal(data, &lease); err != nil {
		return OwnerLease{}, err
	}
	return lease, nil
}

func (s *OwnerStore) Save(lease OwnerLease) error {
	data, err := json.MarshalIndent(lease, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(s.layout.OwnerPath, data, 0o644)
}

func (s *OwnerStore) Clear(leaseID string) error {
	current, err := s.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if current.ID != leaseID {
		return nil
	}
	if err := os.Remove(s.layout.OwnerPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func (s *OwnerStore) Acquire(jobs *JobStore, purpose string) (*OwnerHandle, error) {
	now := time.Now().UTC()
	current, err := s.Load()
	switch {
	case err == nil:
		stale, err := ownerLeaseStale(current, now)
		if err != nil {
			return nil, err
		}
		if !stale {
			return nil, errOwnerActive
		}
		if err := jobs.RecoverInterrupted(); err != nil {
			return nil, err
		}
	case errors.Is(err, os.ErrNotExist):
	default:
		return nil, err
	}

	lease := OwnerLease{
		ID:          newJobID(),
		PID:         os.Getpid(),
		Purpose:     purpose,
		AcquiredAt:  now.Format(time.RFC3339),
		HeartbeatAt: now.Format(time.RFC3339),
	}
	if err := s.Save(lease); err != nil {
		return nil, err
	}
	handle := &OwnerHandle{
		store:  s,
		lease:  lease,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
	go handle.heartbeat()
	return handle, nil
}

func (h *OwnerHandle) Release() error {
	var err error
	h.once.Do(func() {
		close(h.stopCh)
		<-h.doneCh
		err = h.store.Clear(h.lease.ID)
	})
	return err
}

func (h *OwnerHandle) heartbeat() {
	ticker := time.NewTicker(ownerHeartbeatStep)
	defer ticker.Stop()
	defer close(h.doneCh)
	for {
		select {
		case <-h.stopCh:
			return
		case tick := <-ticker.C:
			h.lease.HeartbeatAt = tick.UTC().Format(time.RFC3339)
			_ = h.store.Save(h.lease)
		}
	}
}

func ownerLeaseStale(lease OwnerLease, now time.Time) (bool, error) {
	if lease.HeartbeatAt == "" {
		return true, nil
	}
	heartbeatAt, err := time.Parse(time.RFC3339, lease.HeartbeatAt)
	if err != nil {
		return false, err
	}
	return now.Sub(heartbeatAt) > ownerLeaseTTL, nil
}

func ownerActiveError(root string) error {
	return errors.New("release panel owner is already active under " + root + "; stop the running server/process or wait for the lease to expire (" + strconv.Itoa(int(ownerLeaseTTL/time.Second)) + "s)")
}
