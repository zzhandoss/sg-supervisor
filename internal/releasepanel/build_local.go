package releasepanel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"

	"sg-supervisor/internal/app"
	"sg-supervisor/internal/release"
)

func (s *Service) StartLocalRelease(ctx context.Context) (Job, error) {
	state, err := s.store.Load()
	if err != nil {
		return Job{}, err
	}
	if err := validateRecipe(state.Recipe); err != nil {
		return Job{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	jobs, err := s.jobs.List()
	if err != nil {
		return Job{}, err
	}
	for _, job := range jobs {
		if job.Type == JobTypeLocalRelease && (job.Status == JobStatusQueued || job.Status == JobStatusRunning) {
			return Job{}, errors.New("a local release job is already running")
		}
	}
	job := Job{
		ID:        newJobID(),
		Type:      JobTypeLocalRelease,
		Status:    JobStatusQueued,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Recipe:    state.Recipe,
		Logs:      []string{"job created"},
	}
	if err := s.jobs.Save(job); err != nil {
		return Job{}, err
	}
	go s.runLocalRelease(context.Background(), job, state)
	return job, nil
}

func (s *Service) runLocalRelease(ctx context.Context, job Job, state State) {
	job.Status = JobStatusRunning
	job.StartedAt = time.Now().UTC().Format(time.RFC3339)
	job.Logs = append(job.Logs, "local release started")
	_ = s.jobs.Save(job)

	report, err := s.buildLocalRelease(ctx, &job, state)
	if err != nil {
		job.Status = JobStatusFailed
		job.Error = err.Error()
		job.Logs = append(job.Logs, "local release failed: "+err.Error())
	} else {
		job.Status = JobStatusSucceeded
		job.Report = &report
		job.Logs = append(job.Logs, "local release finished")
	}
	job.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	_ = s.jobs.Save(job)
}

func (s *Service) buildLocalRelease(ctx context.Context, job *Job, state State) (release.SetReport, error) {
	versionDir := filepath.Join(s.layout.ReleasesDir, "v"+state.Recipe.InstallerVersion)
	if _, err := os.Stat(versionDir); err == nil {
		return release.SetReport{}, errors.New("release version already exists: " + versionDir)
	}
	platform, err := hostPlatform()
	if err != nil {
		return release.SetReport{}, err
	}
	job.Logs = append(job.Logs, "building local installer for host platform "+platform)
	_ = s.jobs.Save(*job)
	report, err := s.buildPlatformRelease(ctx, job, state, platform)
	if err != nil {
		return release.SetReport{}, err
	}
	return release.BuildSet(s.layout.Root, state.Recipe.InstallerVersion, []release.Report{report})
}

func (s *Service) buildPlatformRelease(ctx context.Context, job *Job, state State, platform string) (release.Report, error) {
	job.Logs = append(job.Logs, "preparing "+platform+" workspace")
	_ = s.jobs.Save(*job)
	workspaceRoot := filepath.Join(s.layout.WorkspacesDir, job.ID, platform)
	if err := os.RemoveAll(workspaceRoot); err != nil {
		return release.Report{}, err
	}
	assets, err := s.downloadAssets(ctx, state, platform)
	if err != nil {
		return release.Report{}, errors.New("asset download failed for " + platform + ": " + err.Error())
	}
	if err := s.core.BuildInstallTree(ctx, state.Recipe, s.layout.CacheDir, workspaceRoot, func(message string) {
		job.Logs = append(job.Logs, message)
		_ = s.jobs.Save(*job)
	}); err != nil {
		return release.Report{}, errors.New("core assembly failed for " + platform + ": " + err.Error())
	}
	if err := prepareWorkspace(workspaceRoot, assets); err != nil {
		return release.Report{}, errors.New("workspace preparation failed for " + platform + ": " + err.Error())
	}
	payloadBundlePath := filepath.Join(workspaceRoot, "payload", payloadArtifactName(trimVersion(state.Recipe.InstallerVersion), platform))
	job.Logs = append(job.Logs, "building local payload bundle")
	_ = s.jobs.Save(*job)
	if _, err := buildLocalPayloadBundle(ctx, workspaceRoot, state, payloadBundlePath); err != nil {
		return release.Report{}, errors.New("payload bundle build failed for " + platform + ": " + err.Error())
	}
	bootstrapRoot := filepath.Join(workspaceRoot, "bootstrap")
	if err := prepareBootstrapWorkspace(bootstrapRoot, platform, state, assets); err != nil {
		return release.Report{}, errors.New("bootstrap workspace preparation failed for " + platform + ": " + err.Error())
	}
	binaryPath := filepath.Join(workspaceRoot, "binaries", supervisorBinaryName(platform))
	if err := s.builder.BuildSupervisor(ctx, state.RepoRoot, platform, binaryPath); err != nil {
		return release.Report{}, errors.New("supervisor build failed for " + platform + ": " + err.Error())
	}
	supervisor, err := app.New(bootstrapRoot)
	if err != nil {
		return release.Report{}, errors.New("workspace bootstrap failed for " + platform + ": " + err.Error())
	}
	job.Logs = append(job.Logs, "building bootstrap installer")
	_ = s.jobs.Save(*job)
	report, err := supervisor.BuildRelease(ctx, platform, state.Recipe.InstallerVersion, binaryPath)
	if err != nil {
		return release.Report{}, errors.New("release build failed for " + platform + ": " + err.Error())
	}
	job.Logs = append(job.Logs, "assembling delivery archive")
	_ = s.jobs.Save(*job)
	return buildDeliveryRelease(ctx, s.layout.Root, state.Recipe.InstallerVersion, platform, report.ArtifactPath, payloadBundlePath, report.Warnings)
}

func supervisorBinaryName(platform string) string {
	if platform == "windows" {
		return "sg-supervisor.exe"
	}
	return "sg-supervisor"
}

func newJobID() string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
