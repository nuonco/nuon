package hooks

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestIsAwaitRunnerHealthyStep(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		event signal.SignalPhaseEvent
		step  *workflowStepRef
		want  bool
	}{
		{
			name: "runner workflow step",
			step: &workflowStepRef{TargetType: string(app.WorkflowStepTargetTypeRunners)},
			want: true,
		},
		{
			name:  "runner health step without enrichment",
			event: signal.SignalPhaseEvent{StepName: "runner healthy"},
			want:  true,
		},
		{
			name: "component workflow step",
			step: &workflowStepRef{TargetType: string(app.WorkflowStepTargetTypeInstallDeploys)},
		},
		{
			name: "missing workflow step",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isAwaitRunnerHealthyStep(tc.event, tc.step); got != tc.want {
				t.Fatalf("isAwaitRunnerHealthyStep() = %v, want %v", got, tc.want)
			}
		})
	}
}
