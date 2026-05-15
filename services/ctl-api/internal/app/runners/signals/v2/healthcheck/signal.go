package healthcheck

import (
	"fmt"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/lib/pq"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

const SignalType signal.SignalType = "healthcheck"

const (
	heartBeatTimeout time.Duration = time.Second * 15
)

type Signal struct {
	RunnerID string `json:"runner_id"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) MaxInFlightAge() time.Duration {
	return 2 * time.Minute
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.RunnerID == "" {
		return errors.New("runner_id is required")
	}

	// Validate runner exists in database
	_, err := activities.AwaitGet(ctx, activities.GetRequest{RunnerID: s.RunnerID})
	if err != nil {
		return errors.Wrap(err, "runner not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get logger")
	}

	// Check if this is a noop health check
	noopHealthCheck, err := s.isNoopHealthCheck(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to check if a noop health check")
	}
	if noopHealthCheck {
		l.Info("skipping health check - runner in noop status")
		return nil
	}

	// Execute the health check
	newStatus, statusChanged, err := s.executeHealthCheck(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to execute health check")
	}

	if statusChanged {
		l.Info("runner status changed",
			zap.String("runner_id", s.RunnerID),
			zap.String("new_status", string(newStatus)),
		)
	}

	return nil
}

func (s *Signal) isNoopHealthCheck(ctx workflow.Context) (bool, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return false, errors.Wrap(err, "unable to get runner")
	}

	// Skip health check for these statuses
	noopStatuses := []app.RunnerStatus{
		app.RunnerStatusPending,
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusOffline,
		app.RunnerStatusAwaitingInstallStackRun,
	}

	for _, status := range noopStatuses {
		if runner.Status == status {
			return true, nil
		}
	}

	return false, nil
}

func (s *Signal) executeHealthCheck(ctx workflow.Context) (app.RunnerStatus, bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to get logger")
	}

	runner, err := activities.AwaitGetByRunnerID(ctx, s.RunnerID)
	if err != nil {
		return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to get runner")
	}

	// Determine new status based on heartbeat
	newStatus := app.RunnerStatusActive
	heartbeat, err := activities.AwaitGetMostRecentHeartBeatRequest(ctx, activities.GetMostRecentHeartBeatRequest{
		RunnerID: s.RunnerID,
		Process:  app.HeartBeatProcessForRunnerGroupType(runner.RunnerGroup.Type),
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newStatus = app.RunnerStatusError
		} else {
			return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to get heartbeat")
		}
	}

	if heartbeat != nil {
		minHeartBeatTS := workflow.Now(ctx).Add(-heartBeatTimeout)
		if heartbeat.CreatedAt.Before(minHeartBeatTS) {
			newStatus = app.RunnerStatusError
		}
	} else {
		newStatus = app.RunnerStatusError
	}

	isChanged := runner.Status != newStatus

	// Create health check record
	_, err = activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID: s.RunnerID,
		Status:   newStatus,
	})
	if err != nil {
		return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to create runner health check")
	}

	// Update runner status if changed
	if isChanged {
		if newStatus != app.RunnerStatusActive {
			l.Error("runner became unhealthy",
				zap.String("runner_id", runner.ID),
				zap.String("org_id", runner.OrgID),
				zap.String("org_name", runner.Org.Name),
				zap.Int("org_priority", runner.Org.Priority),
			)
		}

		// Update legacy runner.Status for observability (not used for control flow)
		_ = activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          s.RunnerID,
			Status:            newStatus,
			StatusDescription: fmt.Sprintf("status change %s -> %s in health check", runner.Status, newStatus),
		})
		statusactivities.AwaitUpdateRunnerStatusV2(ctx, statusactivities.UpdateRunnerStatusV2Request{
			RunnerID:          s.RunnerID,
			Status:            newStatus,
			StatusDescription: fmt.Sprintf("status change %s -> %s in health check", runner.Status, newStatus),
		})
	}

	// Compute and update warnings
	warnings, isAliasTag := s.computeWarnings(runner, heartbeat)
	if err := activities.AwaitUpdateWarnings(ctx, activities.UpdateWarningsRequest{
		RunnerID:   s.RunnerID,
		Warnings:   warnings,
		IsAliasTag: isAliasTag,
	}); err != nil {
		l.Warn("unable to update runner warnings", zap.Error(err))
	}

	return newStatus, isChanged, nil
}

func (s *Signal) computeWarnings(runner *app.Runner, heartbeat *app.RunnerHeartBeat) (pq.StringArray, bool) {
	var warnings pq.StringArray

	if heartbeat == nil {
		return warnings, false
	}

	expectedVersion := runner.RunnerGroup.Settings.ContainerImageTag
	reportedVersion := heartbeat.Version

	// If the configured tag is not a semver (e.g. "latest", "stable"), it's an alias.
	// Don't show a version mismatch warning for alias tags since the runner reports
	// the resolved version, not the alias name.
	if expectedVersion != "" {
		if _, err := semver.NewVersion(expectedVersion); err != nil {
			// Not a semver — this is an alias tag.
			return warnings, true
		}
	}

	if expectedVersion != "" && reportedVersion != "" && expectedVersion != reportedVersion {
		warnings = append(warnings, fmt.Sprintf("Reported version (%s) does not match configured version (%s).", reportedVersion, expectedVersion))
	}

	return warnings, false
}
