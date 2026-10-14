package componenthelmrecover

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const (
	checkpointLogStream = "log-stream"
	checkpointRecover   = "recover"
)

type logStreamCheckpoint struct {
	LogStreamID string `json:"log_stream_id,omitempty"`
}

type phaseCheckpoint struct {
	JobID     string `json:"job_id,omitempty"`
	Finalized bool   `json:"finalized,omitempty"`
}

func activeCheckpointJobID(ctx workflow.Context) string {
	var cp phaseCheckpoint
	ok, err := job.LoadCheckpoint(ctx, checkpointRecover, &cp)
	if err != nil || !ok || cp.Finalized {
		return ""
	}
	return cp.JobID
}
