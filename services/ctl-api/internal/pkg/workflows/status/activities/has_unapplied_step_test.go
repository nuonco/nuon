package statusactivities

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func step(idx, groupIdx int, groupID string, status app.Status, retried bool) app.WorkflowStep {
	return app.WorkflowStep{
		Idx:                 idx,
		GroupIdx:            groupIdx,
		WorkflowStepGroupID: groupID,
		Status:              app.CompositeStatus{Status: status},
		Retried:             retried,
	}
}

func TestHasUnappliedStep(t *testing.T) {
	tests := []struct {
		name  string
		steps []app.WorkflowStep
		want  bool
	}{
		{
			name: "all success",
			steps: []app.WorkflowStep{
				step(100, 1, "", app.StatusSuccess, false),
				step(200, 2, "", app.StatusSuccess, false),
			},
		},
		{
			name: "noop plan discards its apply",
			steps: []app.WorkflowStep{
				step(400, 4, "", app.StatusAutoSkipped, false),
				step(500, 4, "", app.StatusDiscarded, false),
				step(600, 5, "", app.StatusSuccess, false),
			},
		},
		{
			name: "noop plan discards its apply, matched by group id",
			steps: []app.WorkflowStep{
				step(400, 4, "wsg1", app.StatusAutoSkipped, false),
				step(500, 0, "wsg1", app.StatusDiscarded, false),
			},
		},
		{
			name: "auto-skip in another group does not cover a discard",
			steps: []app.WorkflowStep{
				step(400, 4, "wsg1", app.StatusAutoSkipped, false),
				step(500, 5, "wsg2", app.StatusDiscarded, false),
			},
			want: true,
		},
		{
			name: "auto-skip after the discarded step does not cover it",
			steps: []app.WorkflowStep{
				step(400, 4, "", app.StatusDiscarded, false),
				step(500, 4, "", app.StatusAutoSkipped, false),
			},
			want: true,
		},
		{
			name: "discard from a stop",
			steps: []app.WorkflowStep{
				step(400, 4, "", app.StatusSuccess, false),
				step(500, 4, "", app.StatusDiscarded, false),
			},
			want: true,
		},
		{
			name: "retried discard",
			steps: []app.WorkflowStep{
				step(500, 4, "", app.StatusDiscarded, true),
				step(501, 4, "", app.StatusSuccess, false),
			},
		},
		{
			name: "retried error followed by success",
			steps: []app.WorkflowStep{
				step(400, 4, "", app.StatusError, true),
				step(401, 4, "", app.StatusError, true),
				step(402, 4, "", app.StatusSuccess, false),
			},
		},
		{
			name: "retried error whose retry also errored",
			steps: []app.WorkflowStep{
				step(400, 4, "", app.StatusError, true),
				step(401, 4, "", app.StatusError, false),
			},
			want: true,
		},
		{
			name: "error",
			steps: []app.WorkflowStep{
				step(400, 4, "", app.StatusAutoSkipped, false),
				step(500, 5, "", app.StatusError, false),
			},
			want: true,
		},
		{
			name: "user skipped",
			steps: []app.WorkflowStep{
				step(500, 5, "", app.StatusUserSkipped, false),
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, hasUnappliedStep(tc.steps))
		})
	}
}
