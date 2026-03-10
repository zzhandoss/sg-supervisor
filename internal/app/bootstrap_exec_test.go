package app

import (
	"context"
	"strings"
	"testing"
)

func TestRunBootstrapCommandIncludesCommandPrefixOnFailure(t *testing.T) {
	output, err := runBootstrapCommand(
		context.Background(),
		t.TempDir(),
		nil,
		"go",
		"test",
		"definitely-not-a-real-package",
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if output == "" {
		t.Fatal("expected command output")
	}
	if !strings.Contains(err.Error(), "go test definitely-not-a-real-package failed") {
		t.Fatalf("unexpected error: %s", err)
	}
}
