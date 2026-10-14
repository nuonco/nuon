package actionworkflowrun

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/imagesync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const (
	checkpointRun         = "run"
	checkpointImageDeps   = "image-deps"
	checkpointImageMirror = "image-mirror"
	checkpointAction      = "action"
)

type runCheckpoint struct {
	RunID       string `json:"run_id,omitempty"`
	LogStreamID string `json:"log_stream_id,omitempty"`
}

type imageDepCheckpoint struct {
	ComponentID        string                   `json:"component_id"`
	BuildID            string                   `json:"build_id"`
	InstallComponentID string                   `json:"install_component_id"`
	Sync               imagesync.SyncCheckpoint `json:"sync"`
}

type imageDepsCheckpoint struct {
	Decided   bool                 `json:"decided,omitempty"`
	Deps      []imageDepCheckpoint `json:"deps,omitempty"`
	Published bool                 `json:"published,omitempty"`
}

type imageMirrorCheckpoint struct {
	JobID     string `json:"job_id,omitempty"`
	DigestRef string `json:"digest_ref,omitempty"`
	Finalized bool   `json:"finalized,omitempty"`
}

type actionCheckpoint struct {
	Plan      *plan.CreateActionPlanResponse `json:"plan,omitempty"`
	JobID     string                         `json:"job_id,omitempty"`
	Finalized bool                           `json:"finalized,omitempty"`
}

func activeCheckpoint(ctx workflow.Context) (string, string) {
	var run runCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointRun, &run); err != nil {
		return "", ""
	}

	var action actionCheckpoint
	if ok, err := job.LoadCheckpoint(ctx, checkpointAction, &action); err == nil && ok && action.JobID != "" {
		if action.Finalized {
			return "", ""
		}
		return action.JobID, run.RunID
	}

	var mirror imageMirrorCheckpoint
	if ok, err := job.LoadCheckpoint(ctx, checkpointImageMirror, &mirror); err == nil && ok && mirror.JobID != "" && !mirror.Finalized {
		return mirror.JobID, run.RunID
	}

	var deps imageDepsCheckpoint
	if ok, err := job.LoadCheckpoint(ctx, checkpointImageDeps, &deps); err == nil && ok {
		for _, dep := range deps.Deps {
			if dep.Sync.JobID != "" && !dep.Sync.Finalized {
				return dep.Sync.JobID, run.RunID
			}
		}
	}

	return "", ""
}
