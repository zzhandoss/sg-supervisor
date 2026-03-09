package app

import (
	"context"
	"testing"
)

func TestStartServiceBlocksAdapterUntilDeviceServiceRuns(t *testing.T) {
	supervisor := newTestApp(t)

	if err := supervisor.StartService(context.Background(), "dahua-terminal-adapter"); err == nil {
		t.Fatalf("expected adapter start to be blocked until device-service runs")
	}
}
