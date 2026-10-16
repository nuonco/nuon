package actionworkflowrun

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const (
	checkpointRun         = "run"
	checkpointImageDeps   = "image-deps"
	checkpointImageMirror = "image-mirror"
	checkpointAction      = "action"
	checkpointActionPlan  = "action-plan"
)

type runCheckpoint struct {
	RunID       string `json:"run_id,omitempty"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}

type imageDepCheckpoint struct {
	ComponentID        string `json:"component_id"`
	BuildID            string `json:"build_id"`
	InstallComponentID string `json:"install_component_id"`
}

func (d imageDepCheckpoint) phaseKey() string {
	return checkpointImageDeps + "-" + d.ComponentID
}

type imageDepsCheckpoint struct {
	Decided   bool                 `json:"decided,omitempty"`
	Deps      []imageDepCheckpoint `json:"deps,omitempty"`
	Published bool                 `json:"published,omitempty"`
}

func activeCheckpoint(ctx workflow.Context) (string, string) {
	var run runCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointRun, &run); err != nil {
		return "", ""
	}

	var action job.Checkpoint
	if ok, err := job.LoadCheckpoint(ctx, checkpointAction, &action); err == nil && ok && action.JobID != "" {
		if action.Finalized {
			return "", ""
		}
		return action.JobID, run.RunID
	}

	var mirror job.Checkpoint
	if ok, err := job.LoadCheckpoint(ctx, checkpointImageMirror, &mirror); err == nil && ok && mirror.JobID != "" && !mirror.Finalized {
		return mirror.JobID, run.RunID
	}

	var deps imageDepsCheckpoint
	if ok, err := job.LoadCheckpoint(ctx, checkpointImageDeps, &deps); err == nil && ok {
		for _, dep := range deps.Deps {
			var cp job.Checkpoint
			if _, err := job.LoadCheckpoint(ctx, dep.phaseKey(), &cp); err == nil && cp.JobID != "" && !cp.Finalized {
				return cp.JobID, run.RunID
			}
		}
	}

	return "", ""
}
