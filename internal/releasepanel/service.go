package releasepanel

import (
	"context"
	"errors"
	"strings"
	"sync"
)

type Service struct {
	layout   Layout
	store    *Store
	jobs     *JobStore
	owner    *OwnerStore
	assets   AssetSource
	node     NodeSource
	builder  BinaryBuilder
	executor Executor
	mu       sync.Mutex
}

func NewService(root, repoRoot string) (*Service, error) {
	layout := NewLayout(root)
	store := NewStore(layout)
	if _, err := store.Ensure(repoRoot); err != nil {
		return nil, err
	}
	jobs := NewJobStore(layout)
	executor := ExecExecutor{}
	return &Service{
		layout:   layout,
		store:    store,
		jobs:     jobs,
		owner:    NewOwnerStore(layout),
		assets:   NewGitHubAssetSource(executor),
		node:     NewNodeDistSource(),
		builder:  NewGoBinaryBuilder(executor),
		executor: executor,
	}, nil
}

func (s *Service) Status(ctx context.Context) (Status, error) {
	if err := ctx.Err(); err != nil {
		return Status{}, err
	}
	platform, err := hostPlatform()
	if err != nil {
		return Status{}, err
	}
	state, err := s.store.Load()
	if err != nil {
		return Status{}, err
	}
	jobs, err := s.jobs.List()
	if err != nil {
		return Status{}, err
	}
	licenses, err := s.store.ListIssuedLicenses()
	if err != nil {
		return Status{}, err
	}
	return Status{
		Root:          s.layout.Root,
		ListenAddress: state.ListenAddress,
		RepoRoot:      state.RepoRoot,
		HostPlatform:  platform,
		Recipe:        state.Recipe,
		Keys: KeysStatus{
			LicensePublicKeyBase64: state.Keys.LicensePublicKeyBase64,
			PackagePublicKeyBase64: state.Keys.PackagePublicKeyBase64,
			LicenseConfigured:      state.Keys.LicensePrivateKeyBase64 != "" && state.Keys.LicensePublicKeyBase64 != "",
			PackageConfigured:      state.Keys.PackagePrivateKeyBase64 != "" && state.Keys.PackagePublicKeyBase64 != "",
		},
		Jobs:           jobs,
		IssuedLicenses: licenses,
		ReleaseDir:     s.layout.ReleasesDir,
	}, nil
}

func (s *Service) UpdateRecipe(ctx context.Context, recipe Recipe) (Status, error) {
	if err := validateRecipe(recipe); err != nil {
		return Status{}, err
	}
	state, err := s.store.Load()
	if err != nil {
		return Status{}, err
	}
	state.Recipe = recipe
	if err := s.store.Save(state); err != nil {
		return Status{}, err
	}
	return s.Status(ctx)
}

func (s *Service) ListVersions(ctx context.Context, repo string) ([]ReleaseVersion, error) {
	switch repo {
	case "school-gate":
		return s.assets.ListVersions(ctx, RepoSchoolGate)
	case "adapter":
		return s.assets.ListVersions(ctx, RepoAdapter)
	case "node":
		return s.node.ListVersions(ctx)
	default:
		return nil, errors.New("unsupported repo")
	}
}

func validateRecipe(recipe Recipe) error {
	if strings.TrimSpace(recipe.InstallerVersion) == "" {
		return errors.New("installerVersion is required")
	}
	if strings.TrimSpace(recipe.SchoolGateVersion) == "" {
		return errors.New("schoolGateVersion is required")
	}
	if strings.TrimSpace(recipe.AdapterVersion) == "" {
		return errors.New("adapterVersion is required")
	}
	if strings.TrimSpace(recipe.NodeVersion) == "" {
		return errors.New("nodeVersion is required")
	}
	return nil
}
