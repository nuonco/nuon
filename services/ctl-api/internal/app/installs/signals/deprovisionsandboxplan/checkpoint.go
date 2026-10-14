package deprovisionsandboxplan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const (
	checkpointRun  = "run"
	checkpointPlan = "plan"
)

type runCheckpoint struct {
	SandboxRunID string `json:"sandbox_run_id,omitempty"`
	LogStreamID  string `json:"log_stream_id,omitempty"`
}

type phaseCheckpoint struct {
	JobID     string `json:"job_id,omitempty"`
	Finalized bool   `json:"finalized,omitempty"`
}

func activeCheckpointJobID(ctx workflow.Context) (string, string) {
	var cp phaseCheckpoint
	ok, err := job.LoadCheckpoint(ctx, checkpointPlan, &cp)
	if err != nil || !ok || cp.Finalized || cp.JobID == "" {
		return "", ""
	}
	var run runCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointRun, &run); err != nil {
		return cp.JobID, ""
	}
	return cp.JobID, run.SandboxRunID
}
