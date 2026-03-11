package app

import (
	"path/filepath"
	"testing"
)

func TestBootstrapBuildStepsIncludeOps(t *testing.T) {
	for _, step := range bootstrapBuildSteps() {
		if step.Name == "build-ops" {
			return
		}
	}
	t.Fatal("expected build-ops step in bootstrap plan")
}

func TestBootstrapDeployTargetsIncludeOps(t *testing.T) {
	for _, target := range bootstrapDeployTargets() {
		if target.Filter == "@school-gate/ops" && target.TargetPath == filepath.Join("packages", "ops") {
			return
		}
	}
	t.Fatal("expected @school-gate/ops deploy target in bootstrap plan")
}
