package updatetag

import (
	"fmt"

	"go.temporal.io/sdk/log"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "update_runner_tag"

type Signal struct {
	RunnerID string `json:"runner_id"`
	Tag      string `json:"tag"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}
	if s.Tag == "" {
		return errors.New("tag is required")
	}

	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	// Update the container image tag in the database
	if err := activities.AwaitUpdateContainerImageTag(ctx, activities.UpdateContainerImageTagRequest{
		RunnerID: s.RunnerID,
		Tag:      s.Tag,
	}); err != nil {
		return errors.Wrap(err, "unable to update container image tag")
	}

	l.Info("updated container image tag", "runner_id", s.RunnerID, "tag", s.Tag)

	// Get runner details to determine type
	runner, err := activities.AwaitGet(ctx, activities.GetRequest{
		RunnerID: s.RunnerID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	if runner.RunnerGroup.Type != app.RunnerGroupTypeInstall {
		return fmt.Errorf("unsupported runner group type: %s", runner.RunnerGroup.Type)
	}
	return s.handleInstallRunner(ctx, l, runner)
}

// handleInstallRunner updates the tag and triggers a graceful shutdown so the runner
// restarts with the new tag.
func (s *Signal) handleInstallRunner(ctx workflow.Context, l log.Logger, runner *app.Runner) error {
	l.Info("install runner: triggering graceful shutdown for tag update", "runner_id", s.RunnerID)

	// Try process-based shutdown first
	process, err := activities.AwaitGetCurrentRunnerProcess(ctx, activities.GetCurrentRunnerProcessRequest{
		RunnerID:    s.RunnerID,
		ProcessType: string(app.RunnerProcessTypeInstall),
	})
	if err == nil && process != nil && process.ID != "" {
		_, err := activities.AwaitCreateRunnerProcessShutdown(ctx, activities.CreateRunnerProcessShutdownRequest{
			RunnerProcessID: process.ID,
			Type:            app.RunnerProcessShutdownTypeGraceful,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create process shutdown")
		}
		return nil
	}

	// Fallback: update status to indicate restart needed
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          s.RunnerID,
		Status:            runner.Status,
		StatusDescription: fmt.Sprintf("tag updated to %s, awaiting restart", s.Tag),
	}); err != nil {
		return errors.Wrap(err, "unable to update runner status")
	}
	statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
		RunnerID:          s.RunnerID,
		Status:            runner.Status,
		StatusDescription: fmt.Sprintf("tag updated to %s, awaiting restart", s.Tag),
	})

	return nil
}
