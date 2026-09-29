package offlinecheck

import (
	"strings"
	"time"

	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const (
	SignalType signal.SignalType = "offline_check"

	offlineDuration time.Duration = time.Hour * 24
)

type Signal struct {
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	_, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	isNoop := generics.SliceContains(runner.Status, []app.RunnerStatus{
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusOffline,
		app.RunnerStatusDisabled,
	})
	if isNoop {
		return nil
	}

	minTS := workflow.Now(ctx).Add(-offlineDuration)
	isOffline := false

	healthCheck, err := activities.AwaitGetHealthCheck(ctx, &activities.GetHealthCheckRequest{
		ID:     s.RunnerID,
		Status: app.RunnerStatusActive,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			if minTS.After(runner.CreatedAt) {
				isOffline = true
			}
		} else {
			return errors.Wrap(err, "unable to get health checks")
		}
	} else {
		if minTS.After(healthCheck.CreatedAt) {
			isOffline = true
		}
	}

	if isOffline {
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusOffline,
			StatusDescription: "runner is offline",
		}); err != nil {
			return errors.Wrap(err, "unable to update runner status")
		}
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            app.RunnerStatusOffline,
			StatusDescription: "runner is offline",
		})
	}

	return nil
}
