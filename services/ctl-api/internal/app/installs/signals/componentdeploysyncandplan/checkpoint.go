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

type phaseCheckpoint struct {
	JobID     string `json:"job_id,omitempty"`
	Finalized bool   `json:"finalized,omitempty"`
}

func activeCheckpointJobID(ctx workflow.Context) string {
	for _, phase := range []string{checkpointPlan, checkpointSync} {
		var cp phaseCheckpoint
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
