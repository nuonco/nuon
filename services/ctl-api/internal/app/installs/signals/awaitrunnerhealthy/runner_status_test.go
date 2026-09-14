package awaitrunnerhealthy

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func TestRunnerCannotBecomeHealthy(t *testing.T) {
	tests := map[app.RunnerStatus]bool{
		app.RunnerStatusOffline:  true,
		app.RunnerStatusError:    true,
		app.RunnerStatusActive:   false,
		app.RunnerStatusPending:  false,
		app.RunnerStatusDisabled: false,
	}

	for status, expected := range tests {
		t.Run(string(status), func(t *testing.T) {
			if actual := runnerCannotBecomeHealthy(status); actual != expected {
				t.Fatalf("expected %t, got %t", expected, actual)
			}
		})
	}
}
