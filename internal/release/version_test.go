package release

import "testing"

func TestNormalizeVersion(t *testing.T) {
	if got := normalizeVersion("v1.2.3"); got != "1.2.3" {
		t.Fatalf("expected v-prefix to be trimmed, got %s", got)
	}
	if got := normalizeVersion("1.2.3"); got != "1.2.3" {
		t.Fatalf("expected plain version to stay unchanged, got %s", got)
	}
}
