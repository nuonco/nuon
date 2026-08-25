package v2

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AppBranchConfigUpdate waits on the runner, then hands off to
// getSandboxReprovisionSteps, which opened with the same wait. Only the stack
// steps run in between, so without them the second gate could only re-confirm
// the first — it showed up as two "runner healthy" steps back to back.
func TestSandboxNeedsRunnerHealthyGate(t *testing.T) {
	tests := []struct {
		name string
		diff *app.InstallConfigDiff
		want bool
	}{
		{
			name: "sandbox only",
			diff: &app.InstallConfigDiff{SandboxChanged: true},
			want: false,
		},
		{
			name: "stack and sandbox",
			diff: &app.InstallConfigDiff{SandboxChanged: true, StackChanged: true},
			want: true,
		},
		{
			name: "no diff",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sandboxNeedsRunnerHealthyGate(tt.diff); got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}
