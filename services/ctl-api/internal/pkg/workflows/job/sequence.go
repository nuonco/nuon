package job

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

type Phase struct {
	Key      string
	Prepare  func(workflow.Context) (string, error)
	Finalize func(workflow.Context, string) error
}

type Checkpoint struct {
	JobID     string `json:"job_id"`
	Finalized bool   `json:"finalized"`
}

type Sequence struct {
	Phases []Phase
	Finish func(workflow.Context, error) error
}

func (s Sequence) Run(ctx workflow.Context) error {
	waiting, err := s.run(ctx)
	if waiting {
		return callback.Suspended()
	}
	if s.Finish != nil {
		return s.Finish(ctx, err)
	}
	return err
}

func (s Sequence) run(ctx workflow.Context) (bool, error) {
	for _, phase := range s.Phases {
		var state Checkpoint
		if _, err := LoadCheckpoint(ctx, phase.Key, &state); err != nil {
			return false, fmt.Errorf("load %s checkpoint: %w", phase.Key, err)
		}
		if state.Finalized {
			continue
		}
		if state.JobID == "" {
			jobID, err := phase.Prepare(ctx)
			if err != nil {
				return false, err
			}
			state.JobID = jobID
			state.Finalized = jobID == ""
			if err := SaveCheckpoint(ctx, phase.Key, state); err != nil {
				return false, fmt.Errorf("checkpoint %s job: %w", phase.Key, err)
			}
			if state.Finalized {
				continue
			}
		}

		status, err := publish(ctx, state.JobID)
		if err != nil {
			return false, err
		}
		if !status.IsTerminal() {
			return true, nil
		}
		if phase.Finalize != nil {
			if err := phase.Finalize(ctx, state.JobID); err != nil {
				return false, err
			}
		}
		state.Finalized = true
		if err := SaveCheckpoint(ctx, phase.Key, state); err != nil {
			return false, fmt.Errorf("checkpoint %s finalization: %w", phase.Key, err)
		}
	}
	return false, nil
}
