package v2

import "testing"

// ReprovisionStack recreates the runner and ends its stack phase waiting on it,
// then hands off to deployAllComponents, which opens with the same wait. Nothing
// runs in between, so the second gate can only re-confirm the first — it showed
// up as a duplicate "runner healthy" step the dashboard rendered as a retry
// attempt of a step that had never run.
func TestNeedsRunnerHealthyGate(t *testing.T) {
	tests := []struct {
		name         string
		lastStepName string
		want         bool
	}{
		{
			name: "nothing emitted yet",
			want: true,
		},
		{
			name:         "previous phase ended on an unrelated step",
			lastStepName: "reprovision sandbox dns if enabled",
			want:         true,
		},
		{
			name:         "previous phase already waited on the runner",
			lastStepName: runnerHealthyStepName,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := newStepGroup(nil)
			sg.lastStepName = tt.lastStepName

			if got := sg.needsRunnerHealthyGate(); got != tt.want {
				t.Fatalf("needsRunnerHealthyGate() = %v, want %v", got, tt.want)
			}
		})
	}
}
