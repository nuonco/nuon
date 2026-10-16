package componentdeploysyncandplan

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const (
	checkpointDeploy = "deploy"
	checkpointSync   = "sync"
	checkpointPlan   = "plan"
)

type deployCheckpoint struct {
	DeployID    string `json:"deploy_id,omitempty"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}

func activeCheckpointJobID(ctx workflow.Context) string {
	for _, phase := range []string{checkpointPlan, checkpointSync} {
		var cp job.Checkpoint
		ok, err := job.LoadCheckpoint(ctx, phase, &cp)
		if err != nil || !ok {
			continue
		}
		if cp.JobID != "" && !cp.Finalized {
			return cp.JobID
		}
	}
	return ""
}
