package reprovisionsandboxapplyplan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const checkpointApply = "apply"

func activeCheckpointJobID(ctx workflow.Context) string {
	var cp job.Checkpoint
	ok, err := job.LoadCheckpoint(ctx, checkpointApply, &cp)
	if err != nil || !ok || cp.Finalized {
		return ""
	}
	return cp.JobID
}
