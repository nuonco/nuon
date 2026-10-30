package helpers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	emitterclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	// MinScheduledReconcileLeadTime is the minimum allowed distance between now
	// and a scheduled reconcile's scheduled_at.
	MinScheduledReconcileLeadTime = time.Minute
	// MaxScheduledReconcileLeadTime is the maximum allowed distance between now
	// and a scheduled reconcile's scheduled_at.
	MaxScheduledReconcileLeadTime = 90 * 24 * time.Hour

	// ScheduledReconcileEmitterPrefix names scheduled-reconcile emitters on the
	// install's action cron signals queue. It is deliberately distinct from the
	// action-cron-/runbook-cron-/drift- prefixes that appconfigupdated.Signal
	// sweeps on that same queue, so a config pin never deletes a pending
	// scheduled reconcile.
	ScheduledReconcileEmitterPrefix = "scheduled-reconcile-"
)

var (
	// ErrScheduledReconcileAlreadyPending is returned when the install already
	// has a pending scheduled reconcile; only one is allowed at a time.
	ErrScheduledReconcileAlreadyPending = errors.New("install already has a pending scheduled reconcile")
	// ErrScheduledReconcileTimeTooSoon is returned when scheduled_at is closer
	// than MinScheduledReconcileLeadTime.
	ErrScheduledReconcileTimeTooSoon = errors.New("scheduled_at must be at least 1 minute in the future")
	// ErrScheduledReconcileTimeTooFar is returned when scheduled_at is further
	// out than MaxScheduledReconcileLeadTime.
	ErrScheduledReconcileTimeTooFar = errors.New("scheduled_at must be at most 90 days in the future")
	// ErrScheduledReconcileNotPending is returned when an operation that
	// requires a pending schedule (e.g. cancel) is attempted on a schedule that
	// has already resolved.
	ErrScheduledReconcileNotPending = errors.New("scheduled reconcile is not pending")
)

// ValidateScheduledReconcileTime enforces the allowed scheduling window.
func ValidateScheduledReconcileTime(scheduledAt time.Time) error {
	now := time.Now()
	if scheduledAt.Before(now.Add(MinScheduledReconcileLeadTime)) {
		return ErrScheduledReconcileTimeTooSoon
	}
	if scheduledAt.After(now.Add(MaxScheduledReconcileLeadTime)) {
		return ErrScheduledReconcileTimeTooFar
	}
	return nil
}

// CreateInstallScheduledReconcileRow validates the scheduling window and the
// one-pending-per-install rule, then creates the schedule row. It does not
// create the backing emitter -- callers create the emitter (whose name is
// derived from the row ID) and then call SetInstallScheduledReconcileEmitterID.
func (h *Helpers) CreateInstallScheduledReconcileRow(ctx context.Context, installID string, scheduledAt time.Time, planOnly bool) (*app.InstallScheduledReconcile, error) {
	if err := ValidateScheduledReconcileTime(scheduledAt); err != nil {
		return nil, err
	}

	var install app.Install
	if err := h.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	var pending int64
	if err := h.db.WithContext(ctx).Model(&app.InstallScheduledReconcile{}).
		Where(app.InstallScheduledReconcile{
			InstallID: installID,
			Status:    app.InstallScheduledReconcileStatusPending,
		}).
		Count(&pending).Error; err != nil {
		return nil, fmt.Errorf("unable to check for pending scheduled reconciles: %w", err)
	}
	if pending > 0 {
		return nil, ErrScheduledReconcileAlreadyPending
	}

	schedule := app.InstallScheduledReconcile{
		InstallID:   installID,
		ScheduledAt: scheduledAt,
		PlanOnly:    planOnly,
		Status:      app.InstallScheduledReconcileStatusPending,
	}
	if err := h.db.WithContext(ctx).Create(&schedule).Error; err != nil {
		// A concurrent request can race past the Count check above; the partial
		// unique index on (install_id) WHERE status = 'pending' is the source of
		// truth and surfaces as a duplicate key error here.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrScheduledReconcileAlreadyPending
		}
		return nil, fmt.Errorf("unable to create scheduled reconcile: %w", err)
	}
	return &schedule, nil
}

// CreateScheduledReconcileEmitter creates the fire-once emitter that dispatches
// a scheduled reconcile, on the install's action cron signals queue.
func (h *Helpers) CreateScheduledReconcileEmitter(
	ctx context.Context,
	installID, scheduleID string,
	scheduledAt time.Time,
	signalType signal.SignalType,
	signalTemplate signal.Signal,
) (*app.QueueEmitter, error) {
	queue, err := h.queueClient.GetQueueByOwnerAndName(ctx, installID, "installs", InstallActionCronSignalsQueueName)
	if err != nil {
		return nil, fmt.Errorf("unable to get install action cron signals queue: %w", err)
	}

	em, err := h.emitterClient.CreateEmitter(ctx, &emitterclient.CreateEmitterRequest{
		QueueID:         queue.ID,
		Name:            ScheduledReconcileEmitterPrefix + scheduleID,
		Description:     fmt.Sprintf("scheduled reconcile %s for install %s", scheduleID, installID),
		Mode:            app.QueueEmitterModeFireOnce,
		ScheduledAt:     &scheduledAt,
		SignalType:      signalType,
		SignalTemplate:  signalTemplate,
		SignalExpiresIn: 15 * time.Minute,
	})
	if err != nil {
		return nil, fmt.Errorf("unable to create scheduled reconcile emitter: %w", err)
	}
	return em, nil
}

// SetInstallScheduledReconcileEmitterID records the backing emitter on the schedule.
func (h *Helpers) SetInstallScheduledReconcileEmitterID(ctx context.Context, scheduleID, emitterID string) error {
	return h.db.WithContext(ctx).
		Model(&app.InstallScheduledReconcile{}).
		Where(app.InstallScheduledReconcile{ID: scheduleID}).
		Update("emitter_id", emitterID).Error
}

// FailInstallScheduledReconcile deletes the emitter (if any) and marks a
// pending schedule failed so it no longer blocks the one-pending-per-install
// rule.
func (h *Helpers) FailInstallScheduledReconcile(ctx context.Context, scheduleID, emitterID, reason string) error {
	if emitterID != "" {
		if err := h.emitterClient.DeleteEmitter(ctx, emitterID); err != nil {
			h.l.Warn("unable to delete scheduled reconcile emitter",
				zap.String("emitter_id", emitterID), zap.Error(err))
		}
	}

	return h.db.WithContext(ctx).
		Model(&app.InstallScheduledReconcile{}).
		Where(app.InstallScheduledReconcile{
			ID:     scheduleID,
			Status: app.InstallScheduledReconcileStatusPending,
		}).
		Updates(map[string]any{
			"status":        app.InstallScheduledReconcileStatusFailed,
			"status_reason": reason,
		}).Error
}

// CancelInstallScheduledReconcile stops and deletes the backing emitter (best
// effort -- a stuck emitter must not block cancelling the schedule) and marks
// the schedule cancelled with reason. Returns ErrScheduledReconcileNotPending
// if the schedule has already resolved.
func (h *Helpers) CancelInstallScheduledReconcile(ctx context.Context, schedule *app.InstallScheduledReconcile, reason string) error {
	if schedule.Status != app.InstallScheduledReconcileStatusPending {
		return ErrScheduledReconcileNotPending
	}

	if schedule.EmitterID != "" {
		if _, err := h.emitterClient.StopEmitter(ctx, schedule.EmitterID); err != nil {
			h.l.Warn("unable to stop scheduled reconcile emitter",
				zap.String("emitter_id", schedule.EmitterID), zap.Error(err))
		}
		if err := h.emitterClient.DeleteEmitter(ctx, schedule.EmitterID); err != nil {
			h.l.Warn("unable to delete scheduled reconcile emitter",
				zap.String("emitter_id", schedule.EmitterID), zap.Error(err))
		}
	}

	res := h.db.WithContext(ctx).
		Model(schedule).
		Updates(map[string]any{
			"status":        app.InstallScheduledReconcileStatusCancelled,
			"status_reason": reason,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to cancel scheduled reconcile: %w", res.Error)
	}
	schedule.Status = app.InstallScheduledReconcileStatusCancelled
	schedule.StatusReason = reason
	return nil
}

// CancelPendingScheduledReconcilesForInstall cancels every pending scheduled
// reconcile for an install (best effort on each row). Used when an install is
// deprovisioned or forgotten.
func (h *Helpers) CancelPendingScheduledReconcilesForInstall(ctx context.Context, installID, reason string) error {
	var schedules []app.InstallScheduledReconcile
	if err := h.db.WithContext(ctx).
		Where(app.InstallScheduledReconcile{
			InstallID: installID,
			Status:    app.InstallScheduledReconcileStatusPending,
		}).
		Find(&schedules).Error; err != nil {
		return fmt.Errorf("unable to get pending scheduled reconciles: %w", err)
	}

	for i := range schedules {
		if err := h.CancelInstallScheduledReconcile(ctx, &schedules[i], reason); err != nil {
			h.l.Warn("unable to cancel scheduled reconcile",
				zap.String("scheduled_reconcile_id", schedules[i].ID), zap.Error(err))
		}
	}
	return nil
}
