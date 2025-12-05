package worker

import (
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	offlineDuration time.Duration = time.Hour * 24
)

// Check if a runner is offline
type OfflineCheckRequest struct {
	RunnerID string `validate:"required" json:"runner_id"`
}

// @temporal-gen workflow
func (w *Workflows) OfflineCheck(ctx workflow.Context, sreq signals.RequestSignal) error {
	startTS := workflow.Now(ctx)
	status := "ok"
	defer func() {
		tags := metrics.ToTags(map[string]string{
			"status": status,
		})
		// write metrics now
		w.mw.Incr(ctx, "runner.health_check", tags...)
		w.mw.Timing(ctx, "runner.health_check.latency", time.Now().Sub(startTS), tags...)
	}()

	err := w.checkOffline(ctx, sreq.ID)
	if err != nil {
		status = "error"
		return errors.Wrap(err, "unable to check if runner was offline")
	}

	return nil
}

func (w *Workflows) checkOffline(ctx workflow.Context, runnerID string) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}

	l = l.With(
		zap.String("runner_id", runner.ID),
		zap.String("org_id", runner.OrgID),
		zap.String("org_name", runner.Org.Name),
		zap.Int("org_priority", runner.Org.Priority),
	)

	isNoop := generics.SliceContains(runner.Status, []app.RunnerStatus{
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusOffline,
	})
	if isNoop {
		return nil
	}

	// last active health check
	minTS := workflow.Now(ctx).Add(-offlineDuration)
	isOffline := false
	healthCheck, err := activities.AwaitGetHealthCheck(ctx, &activities.GetHealthCheckRequest{
		ID:     runnerID,
		Status: app.RunnerStatusActive,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(err.Error(), "not found") {
			if minTS.After(runner.CreatedAt) {
				l.Error("runner is offline because no successful health check was ever found")
				isOffline = true
			}
		} else {
			return errors.Wrap(err, "unable to get health checks")
		}
	} else {
		if minTS.After(healthCheck.CreatedAt) {
			l.Error("runner is offline because most recent health check was too long ago")
			isOffline = true
		}
	}

	if isOffline {
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          runnerID,
			Status:            app.RunnerStatusOffline,
			StatusDescription: fmt.Sprintf("runner is offline"),
		}); err != nil {
			return errors.Wrap(err, "unable to update runner status")
		}
	}

	return nil
}
