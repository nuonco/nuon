package syncsecrets

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const (
	checkpointLogStream   = "log-stream"
	checkpointSyncSecrets = "sync-secrets"
)

type logStreamCheckpoint struct {
	LogStreamID string `json:"log_stream_id,omitempty"`
}

func activeCheckpointJobID(ctx workflow.Context) string {
	var cp job.Checkpoint
	ok, err := job.LoadCheckpoint(ctx, checkpointSyncSecrets, &cp)
	if err != nil || !ok || cp.Finalized {
		return ""
	}
	return cp.JobID
}
