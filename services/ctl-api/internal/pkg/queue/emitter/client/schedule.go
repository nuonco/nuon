package client

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
)

// scheduleWorkflowIDForEmitter is the base workflow ID for schedule-triggered emits.
// Temporal appends a nominal-time suffix per scheduled action.
func scheduleWorkflowIDForEmitter(em *app.QueueEmitter) string {
	return fmt.Sprintf("queue-schedule-emit-%s", em.ID)
}

// scheduleOptionsForEmitter maps an emitter row onto Temporal Schedule options.
// The schedule ID is the emitter Name (stable + deterministic from the reconciled
// desired set, mirroring the legacy emitter-name prefixes). Overlap SKIP replaces
// the in-flight dedup; Jitter maps JitterWindow.
func scheduleOptionsForEmitter(em *app.QueueEmitter) tclient.ScheduleOptions {
	return tclient.ScheduleOptions{
		ID: em.Name,
		Spec: tclient.ScheduleSpec{
			CronExpressions: []string{em.CronSchedule},
			Jitter:          em.JitterWindow,
		},
		Action: &tclient.ScheduleWorkflowAction{
			ID:        scheduleWorkflowIDForEmitter(em),
			Workflow:  "ScheduleEmit",
			TaskQueue: workflows.APITaskQueue,
			Args: []any{emitter.ScheduleEmitRequest{
				EmitterID: em.ID,
				QueueID:   em.QueueID,
			}},
			RetryPolicy: &temporal.RetryPolicy{MaximumAttempts: 0},
			Memo:        emitterMemo(em),
		},
		Overlap: enumsv1.SCHEDULE_OVERLAP_POLICY_SKIP,
	}
}

// EnsureSchedule creates (or updates in place) the Temporal Schedule backing a cron
// emitter. Idempotent: safe to call on every reconcile.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) EnsureSchedule(ctx context.Context, emitterID string) error {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return errors.Wrap(err, "unable to get emitter")
	}
	if em.Mode != app.QueueEmitterModeCron {
		return errors.Errorf("schedules only supported for cron emitters, got %s", em.Mode)
	}

	sc, err := c.tClient.ScheduleClientInNamespace(em.Workflow.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get schedule client")
	}

	opts := scheduleOptionsForEmitter(em)
	if _, err := sc.Create(ctx, opts); err == nil {
		c.l.Debug("schedule created", zap.String("schedule-id", opts.ID), zap.String("emitter-id", emitterID))
		return nil
	} else if !errors.Is(err, temporal.ErrScheduleAlreadyRunning) {
		return errors.Wrap(err, "unable to create schedule")
	}

	// Schedule already exists - replace its spec + action in place.
	handle := sc.GetHandle(ctx, opts.ID)
	if err := handle.Update(ctx, tclient.ScheduleUpdateOptions{
		DoUpdate: func(in tclient.ScheduleUpdateInput) (*tclient.ScheduleUpdate, error) {
			in.Description.Schedule.Spec = &opts.Spec
			in.Description.Schedule.Action = opts.Action
			return &tclient.ScheduleUpdate{Schedule: &in.Description.Schedule}, nil
		},
	}); err != nil {
		return errors.Wrap(err, "unable to update schedule")
	}

	c.l.Debug("schedule updated", zap.String("schedule-id", opts.ID), zap.String("emitter-id", emitterID))
	return nil
}

type NativeSchedulingEnabledRequest struct{}

// NativeSchedulingEnabled exposes the global native-scheduling flag to workflows
// (reconcilers) that cannot read Config directly. Read at the work layer so the
// flip is instant and reversible.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (c *Client) NativeSchedulingEnabled(ctx context.Context, _ *NativeSchedulingEnabledRequest) (bool, error) {
	return c.cfg.NativeSchedulingEnabled, nil
}

type DeleteScheduleRequest struct {
	ScheduleID string `validate:"required"`
	Namespace  string `validate:"required"`
}

// DeleteSchedule removes a Temporal Schedule by ID. Tolerates not-found so it is
// idempotent (a stale schedule for a deleted emitter is harmless: ScheduleEmit
// no-ops when the emitter row is gone).
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (c *Client) DeleteSchedule(ctx context.Context, req *DeleteScheduleRequest) error {
	sc, err := c.tClient.ScheduleClientInNamespace(req.Namespace)
	if err != nil {
		return errors.Wrap(err, "unable to get schedule client")
	}

	handle := sc.GetHandle(ctx, req.ScheduleID)
	if err := handle.Delete(ctx); err != nil {
		c.l.Debug("schedule delete tolerated", zap.String("schedule-id", req.ScheduleID), zap.Error(err))
	}

	return nil
}
