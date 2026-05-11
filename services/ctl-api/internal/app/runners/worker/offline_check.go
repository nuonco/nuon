package worker

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const (
	offlineDuration time.Duration = time.Hour * 24
)

// Check if a runner is offline
type OfflineCheckRequest struct {
	RunnerID string `validate:"required" json:"runner_id"`
}

// @temporal-gen-v2 workflow
func (w *Workflows) OfflineCheck(ctx workflow.Context, sreq signals.RequestSignal) (err error) {
	startTS := workflow.Now(ctx)
	var ownerLabels map[string]string
	defer func() {
		status := "ok"
		if err != nil {
			status = "error"
		}
		tagMap := make(map[string]string, len(ownerLabels)+1)
		for k, v := range ownerLabels {
			tagMap[k] = v
		}
		tagMap["status"] = status
		tags := metrics.ToTags(tagMap)
		// write metrics now
		w.mw.Incr(ctx, "runner.health_check", tags...)
		w.mw.Timing(ctx, "runner.health_check.latency", time.Since(startTS), tags...)
	}()

	runner, err := activities.AwaitGetByRunnerID(ctx, sreq.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner")
	}
	switch runner.RunnerGroup.OwnerType {
	case "installs":
		install, instErr := installactivities.AwaitGetByInstallID(ctx, runner.RunnerGroup.OwnerID)
		if instErr != nil {
			return errors.Wrap(instErr, "unable to get install for runner")
		}
		ownerLabels = install.Labels
	case "orgs":
		ownerLabels = runner.Org.Labels
	}

	if err := w.checkOffline(ctx, runner); err != nil {
		return errors.Wrap(err, "unable to check if runner was offline")
	}

	return nil
}

func (w *Workflows) checkOffline(ctx workflow.Context, runner *app.Runner) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
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
		ID:     runner.ID,
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
			RunnerID:          runner.ID,
			Status:            app.RunnerStatusOffline,
			StatusDescription: "runner is offline",
		}); err != nil {
			return errors.Wrap(err, "unable to update runner status")
		}
	}

	return nil
}
