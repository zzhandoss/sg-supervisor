package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"sg-supervisor/internal/releasepanel"
	"sg-supervisor/internal/releasepanelhttp"
)

func main() {
	if err := run(context.Background(), os.Args[1:]); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	args = normalizeArgs(args)
	switch args[0] {
	case "serve":
		return runServe(ctx, args[1:])
	case "status":
		return runStatus(ctx, args[1:])
	case "set-recipe":
		return runSetRecipe(ctx, args[1:])
	case "list-versions":
		return runListVersions(ctx, args[1:])
	case "build-local-release":
		return runBuildLocalRelease(ctx, args[1:])
	case "issue-license":
		return runIssueLicense(ctx, args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runServe(ctx context.Context, args []string) error {
	service, state, err := newService(args)
	if err != nil {
		return err
	}
	owner, err := service.AcquireOwner("serve")
	if err != nil {
		return err
	}
	defer owner.Release()
	fmt.Printf("Release Panel started on http://%s\n", state.ListenAddress)
	server := releasepanelhttp.NewServer(state.ListenAddress, service)
	serverCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	return server.Run(serverCtx)
}

func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return []string{"serve"}
	}
	if len(args) > 0 && len(args[0]) > 0 && args[0][0] == '-' {
		return append([]string{"serve"}, args...)
	}
	return args
}

func runStatus(ctx context.Context, args []string) error {
	service, _, err := newService(args)
	if err != nil {
		return err
	}
	status, err := service.Status(ctx)
	if err != nil {
		return err
	}
	return printJSON(status)
}

func runSetRecipe(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("set-recipe", flag.ContinueOnError)
	root := fs.String("root", ".release-panel", "release panel root")
	repoRoot := fs.String("repo-root", ".", "repository root")
	installerVersion := fs.String("installer-version", "", "installer version")
	schoolGateVersion := fs.String("school-gate-version", "", "school-gate version")
	adapterVersion := fs.String("adapter-version", "", "adapter version")
	nodeVersion := fs.String("node-version", "", "node runtime version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	service, err := releasepanel.NewService(filepath.Clean(*root), filepath.Clean(*repoRoot))
	if err != nil {
		return err
	}
	status, err := service.UpdateRecipe(ctx, releasepanel.Recipe{
		InstallerVersion:  *installerVersion,
		SchoolGateVersion: *schoolGateVersion,
		AdapterVersion:    *adapterVersion,
		NodeVersion:       *nodeVersion,
	})
	if err != nil {
		return err
	}
	return printJSON(status)
}

func runListVersions(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("list-versions", flag.ContinueOnError)
	root := fs.String("root", ".release-panel", "release panel root")
	repoRoot := fs.String("repo-root", ".", "repository root")
	repo := fs.String("repo", "", "one of: school-gate, adapter, node")
	if err := fs.Parse(args); err != nil {
		return err
	}
	service, err := releasepanel.NewService(filepath.Clean(*root), filepath.Clean(*repoRoot))
	if err != nil {
		return err
	}
	versions, err := service.ListVersions(ctx, *repo)
	if err != nil {
		return err
	}
	return printJSON(versions)
}

func runBuildLocalRelease(ctx context.Context, args []string) error {
	service, _, err := newService(args)
	if err != nil {
		return err
	}
	owner, err := service.AcquireOwner("build-local-release")
	if err != nil {
		return err
	}
	defer owner.Release()
	job, err := service.StartLocalRelease(ctx)
	if err != nil {
		return err
	}
	finalJob, err := waitForJob(ctx, service, job.ID)
	if err != nil {
		return err
	}
	return printJSON(finalJob)
}

func runIssueLicense(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("issue-license", flag.ContinueOnError)
	root := fs.String("root", ".release-panel", "release panel root")
	repoRoot := fs.String("repo-root", ".", "repository root")
	activationRequest := fs.String("activation-request", "", "activation request path")
	customer := fs.String("customer", "", "customer")
	mode := fs.String("mode", "free", "free or bound")
	edition := fs.String("edition", "standard", "edition")
	features := fs.String("features", "", "comma separated features")
	expiresAt := fs.String("expires-at", "", "RFC3339 expiration")
	fingerprint := fs.String("fingerprint", "", "hardware fingerprint for bound mode")
	perpetual := fs.Bool("perpetual", false, "issue perpetual license")
	if err := fs.Parse(args); err != nil {
		return err
	}
	service, err := releasepanel.NewService(filepath.Clean(*root), filepath.Clean(*repoRoot))
	if err != nil {
		return err
	}
	record, err := service.IssueLicense(ctx, releasepanel.LicenseIssueRequest{
		ActivationRequestPath: *activationRequest,
		Customer:              *customer,
		Mode:                  *mode,
		Edition:               *edition,
		Features:              splitCSV(*features),
		ExpiresAt:             *expiresAt,
		Fingerprint:           *fingerprint,
		Perpetual:             *perpetual,
	})
	if err != nil {
		return err
	}
	return printJSON(record)
}

func newService(args []string) (*releasepanel.Service, releasepanel.State, error) {
	fs := flag.NewFlagSet("release-panel", flag.ContinueOnError)
	root := fs.String("root", ".release-panel", "release panel root")
	repoRoot := fs.String("repo-root", ".", "repository root")
	if err := fs.Parse(args); err != nil {
		return nil, releasepanel.State{}, err
	}
	service, err := releasepanel.NewService(filepath.Clean(*root), filepath.Clean(*repoRoot))
	if err != nil {
		return nil, releasepanel.State{}, err
	}
	status, err := service.Status(context.Background())
	if err != nil {
		return nil, releasepanel.State{}, err
	}
	return service, releasepanel.State{
		ListenAddress: status.ListenAddress,
		RepoRoot:      status.RepoRoot,
		Recipe:        status.Recipe,
	}, nil
}

func printJSON(value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	result := make([]string, 0, 8)
	current := ""
	for _, runeValue := range value {
		if runeValue == ',' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
			continue
		}
		if runeValue != ' ' {
			current += string(runeValue)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func waitForJob(ctx context.Context, service *releasepanel.Service, jobID string) (releasepanel.Job, error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		status, err := service.Status(ctx)
		if err != nil {
			return releasepanel.Job{}, err
		}
		for _, job := range status.Jobs {
			if job.ID != jobID {
				continue
			}
			if job.Status == releasepanel.JobStatusQueued || job.Status == releasepanel.JobStatusRunning {
				break
			}
			if job.Status == releasepanel.JobStatusFailed {
				return job, fmt.Errorf("%s", job.Error)
			}
			return job, nil
		}
		select {
		case <-ctx.Done():
			return releasepanel.Job{}, ctx.Err()
		case <-ticker.C:
		}
	}
}
