package releasepanel

import "errors"

func (s *Service) AcquireOwner(purpose string) (*OwnerHandle, error) {
	handle, err := s.owner.Acquire(s.jobs, purpose)
	if err != nil {
		if errors.Is(err, errOwnerActive) {
			return nil, ownerActiveError(s.layout.Root)
		}
		return nil, err
	}
	return handle, nil
}
