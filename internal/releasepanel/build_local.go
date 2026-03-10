package releasepanel

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"time"
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

func (s *Service) buildLocalRelease(ctx context.Context, job *Job, state State) (LocalReleaseSetReport, error) {
	versionDir := filepath.Join(s.layout.ReleasesDir, "v"+state.Recipe.InstallerVersion)
	if _, err := os.Stat(versionDir); err == nil {
		return LocalReleaseSetReport{}, errors.New("release version already exists: " + versionDir)
	}
	platform, err := hostPlatform()
	if err != nil {
		return LocalReleaseSetReport{}, err
	}
	job.Logs = append(job.Logs, "building local installer for host platform "+platform)
	_ = s.jobs.Save(*job)
	report, err := s.buildPlatformRelease(ctx, job, state, platform)
	if err != nil {
		return LocalReleaseSetReport{}, err
	}
	return buildLocalReleaseSet(s.layout.Root, state.Recipe.InstallerVersion, []LocalReleaseReport{report})
}

func (s *Service) buildPlatformRelease(ctx context.Context, job *Job, state State, platform string) (LocalReleaseReport, error) {
	job.Logs = append(job.Logs, "preparing "+platform+" workspace")
	_ = s.jobs.Save(*job)
	workspaceRoot := filepath.Join(s.layout.WorkspacesDir, job.ID, platform)
	if err := os.RemoveAll(workspaceRoot); err != nil {
		return LocalReleaseReport{}, err
	}
	assets, err := s.downloadAssets(ctx, state, platform)
	if err != nil {
		return LocalReleaseReport{}, errors.New("asset download failed for " + platform + ": " + err.Error())
	}
	bootstrapRoot := filepath.Join(workspaceRoot, "bootstrap")
	if err := prepareBootstrapWorkspace(bootstrapRoot, platform, state, assets); err != nil {
		return LocalReleaseReport{}, errors.New("bootstrap workspace preparation failed for " + platform + ": " + err.Error())
	}
	binaryPath := filepath.Join(workspaceRoot, "binaries", supervisorBinaryName(platform))
	if err := s.builder.BuildSupervisor(ctx, state.RepoRoot, platform, binaryPath); err != nil {
		return LocalReleaseReport{}, errors.New("supervisor build failed for " + platform + ": " + err.Error())
	}
	targetBinaryPath := filepath.Join(bootstrapRoot, filepath.Base(binaryPath))
	if err := copyArtifact(binaryPath, targetBinaryPath); err != nil {
		return LocalReleaseReport{}, errors.New("bootstrap workspace preparation failed for " + platform + ": " + err.Error())
	}
	job.Logs = append(job.Logs, "building bootstrap package")
	_ = s.jobs.Save(*job)
	job.Logs = append(job.Logs, "assembling delivery archive")
	_ = s.jobs.Save(*job)
	freeLicensePath, err := s.latestIssuedFreeLicensePath()
	if err != nil {
		return LocalReleaseReport{}, errors.New("license lookup failed: " + err.Error())
	}
	if freeLicensePath != "" {
		job.Logs = append(job.Logs, "bundling latest free license into delivery archive")
		_ = s.jobs.Save(*job)
	}
	deliveryRoot, err := prepareDeliveryRoot(bootstrapRoot, state, assets, freeLicensePath)
	if err != nil {
		return LocalReleaseReport{}, errors.New("delivery root preparation failed for " + platform + ": " + err.Error())
	}
	return buildDeliveryRelease(ctx, s.layout.Root, state.Recipe.InstallerVersion, platform, deliveryRoot, nil)
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
