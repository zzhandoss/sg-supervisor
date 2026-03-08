package releasepanel

import (
	"context"
	"os"
	"path/filepath"
)

type BinaryBuilder interface {
	BuildSupervisor(ctx context.Context, repoRoot, platform, outputPath string) error
}

type GoBinaryBuilder struct {
	executor Executor
}

func NewGoBinaryBuilder(executor Executor) *GoBinaryBuilder {
	return &GoBinaryBuilder{executor: executor}
}

func (b *GoBinaryBuilder) BuildSupervisor(ctx context.Context, repoRoot, platform, outputPath string) error {
	env := map[string]string{
		"GOARCH":      "amd64",
		"CGO_ENABLED": "0",
	}
	switch platform {
	case "windows":
		env["GOOS"] = "windows"
	case "linux":
		env["GOOS"] = "linux"
	default:
		return os.ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	_, err := b.executor.Run(ctx, repoRoot, env, "go", "build", "-o", outputPath, "./cmd/sg-supervisor")
	return err
}
