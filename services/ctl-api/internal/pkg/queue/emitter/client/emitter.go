package client

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	enumsv1 "go.temporal.io/api/enums/v1"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/nuonco/nuon/pkg/workflows"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/emitter"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const (
	defaultEmitterWorkflowIDTemplate string = "queue-emitter-%s"
)

type CreateEmitterRequest struct {
	QueueID     string `validate:"required"`
	Name        string `validate:"required"`
	Description string

	// Mode determines if this is a recurring cron emitter or a one-shot scheduled emitter
	Mode app.QueueEmitterMode `validate:"required"`

	// For cron mode: the cron schedule expression (e.g., "0 * * * *")
	CronSchedule string
	// For scheduled mode: when to fire the signal
	ScheduledAt *time.Time

	SignalType     signal.SignalType `validate:"required"`
	SignalTemplate signal.Signal
}

func (c *Client) CreateEmitter(ctx context.Context, req *CreateEmitterRequest) (*app.QueueEmitter, error) {
	switch req.Mode {
	case app.QueueEmitterModeCron:
		if req.CronSchedule == "" {
			return nil, errors.New("cron_schedule is required for cron mode")
		}
	case app.QueueEmitterModeScheduled:
		if req.ScheduledAt == nil {
			return nil, errors.New("scheduled_at is required for scheduled mode")
		}
	default:
		return nil, errors.Errorf("invalid emitter mode: %s", req.Mode)
	}

	q, err := c.getQueue(ctx, req.QueueID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get queue")
	}

	em := app.QueueEmitter{
		QueueID:      q.ID,
		Name:         req.Name,
		Description:  req.Description,
		Mode:         req.Mode,
		CronSchedule: req.CronSchedule,
		ScheduledAt:  req.ScheduledAt,
		SignalType:   req.SignalType,
		SignalTemplate: signaldb.SignalData{
			Signal: req.SignalTemplate,
		},
		Status: app.NewCompositeStatus(ctx, app.StatusPending),
		Workflow: signaldb.WorkflowRef{
			Namespace:  q.Workflow.Namespace,
			IDTemplate: defaultEmitterWorkflowIDTemplate,
		},
	}

	if res := c.db.WithContext(ctx).Create(&em); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to create emitter")
	}

	wkflowReq := emitter.EmitterWorkflowRequest{
		EmitterID: em.ID,
		Version:   c.cfg.Version,
	}

	opts := tclient.StartWorkflowOptions{
		ID:        em.Workflow.ID,
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"id":       em.ID,
			"queue-id": q.ID,
			"name":     em.Name,
			"mode":     string(em.Mode),
			"emitter":  true,
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	if req.Mode == app.QueueEmitterModeCron {
		opts.CronSchedule = req.CronSchedule
	}

	wkflowRun, err := c.tClient.ExecuteWorkflowInNamespace(ctx,
		em.Workflow.Namespace,
		opts,
		"Emitter",
		wkflowReq,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to start emitter workflow")
	}

	c.l.Debug("emitter workflow started",
		zap.String("id", em.ID),
		zap.String("queue-id", q.ID),
		zap.String("workflow-id", em.Workflow.ID),
		zap.String("run-id", wkflowRun.GetRunID()),
		zap.String("mode", string(em.Mode)),
	)

	em.Status = app.NewCompositeStatus(ctx, app.StatusInProgress)
	if res := c.db.WithContext(ctx).Save(&em); res.Error != nil {
		c.l.Warn("failed to update emitter status", zap.Error(res.Error))
	}

	return &em, nil
}

func (c *Client) GetEmitter(ctx context.Context, emitterID string) (*app.QueueEmitter, error) {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get emitter")
	}
	return em, nil
}

func (c *Client) getEmitter(ctx context.Context, emitterID string) (*app.QueueEmitter, error) {
	var em app.QueueEmitter
	if res := c.db.WithContext(ctx).First(&em, "id = ?", emitterID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitter")
	}
	return &em, nil
}

func (c *Client) GetEmittersByQueueID(ctx context.Context, queueID string) ([]app.QueueEmitter, error) {
	var emitters []app.QueueEmitter
	if res := c.db.WithContext(ctx).Where("queue_id = ?", queueID).Find(&emitters); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get emitters")
	}
	return emitters, nil
}

func (c *Client) PauseEmitter(ctx context.Context, emitterID string) (*app.QueueEmitter, error) {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get emitter")
	}

	em.Status = app.NewCompositeStatus(ctx, app.StatusCancelled)
	if res := c.db.WithContext(ctx).Save(em); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update emitter status")
	}

	c.l.Debug("emitter paused", zap.String("id", emitterID))
	return em, nil
}

func (c *Client) ResumeEmitter(ctx context.Context, emitterID string) (*app.QueueEmitter, error) {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get emitter")
	}

	em.Status = app.NewCompositeStatus(ctx, app.StatusInProgress)
	if res := c.db.WithContext(ctx).Save(em); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to update emitter status")
	}

	c.l.Debug("emitter resumed", zap.String("id", emitterID))
	return em, nil
}

func (c *Client) DeleteEmitter(ctx context.Context, emitterID string) error {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return errors.Wrap(err, "unable to get emitter")
	}

	if res := c.db.WithContext(ctx).Delete(em); res.Error != nil {
		return errors.Wrap(res.Error, "unable to delete emitter")
	}

	c.l.Debug("emitter deleted", zap.String("id", emitterID))
	return nil
}

func (c *Client) EnsureRunning(ctx context.Context, emitterID string) (*emitter.EnsureRunningResponse, error) {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get emitter")
	}

	rawResp, err := c.tClient.UpdateWorkflowInNamespace(ctx, em.Workflow.Namespace, tclient.UpdateWorkflowOptions{
		WorkflowID:   em.Workflow.ID,
		UpdateName:   emitter.EnsureRunningUpdateName,
		WaitForStage: tclient.WorkflowUpdateStageCompleted,
		Args: []any{
			emitter.EnsureRunningRequest{},
		},
	})
	if err != nil {
		return nil, errors.Wrap(err, "emitter workflow is not running or unreachable")
	}

	var resp emitter.EnsureRunningResponse
	if err := rawResp.Get(ctx, &resp); err != nil {
		return nil, errors.Wrap(err, "unable to get response")
	}

	c.l.Debug("emitter is running",
		zap.String("id", emitterID),
		zap.String("mode", resp.Mode),
		zap.Int64("emit-count", resp.EmitCount),
	)

	return &resp, nil
}

func (c *Client) RestartEmitterWorkflow(ctx context.Context, emitterID string) (*app.QueueEmitter, error) {
	em, err := c.getEmitter(ctx, emitterID)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get emitter")
	}

	if em.Mode == app.QueueEmitterModeScheduled && em.Fired {
		return nil, errors.New("cannot restart a scheduled emitter that has already fired")
	}

	wkflowReq := emitter.EmitterWorkflowRequest{
		EmitterID: em.ID,
		Version:   c.cfg.Version,
	}

	opts := tclient.StartWorkflowOptions{
		ID:        em.Workflow.ID,
		TaskQueue: workflows.APITaskQueue,
		Memo: map[string]any{
			"id":       em.ID,
			"queue-id": em.QueueID,
			"name":     em.Name,
			"mode":     string(em.Mode),
			"emitter":  true,
		},
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 0,
		},
	}

	if em.Mode == app.QueueEmitterModeCron {
		opts.CronSchedule = em.CronSchedule
	}

	wkflowRun, err := c.tClient.ExecuteWorkflowInNamespace(ctx,
		em.Workflow.Namespace,
		opts,
		"Emitter",
		wkflowReq,
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to restart emitter workflow")
	}

	c.l.Debug("emitter workflow restarted",
		zap.String("id", em.ID),
		zap.String("workflow-id", em.Workflow.ID),
		zap.String("run-id", wkflowRun.GetRunID()),
	)

	em.Status = app.NewCompositeStatus(ctx, app.StatusInProgress)
	if res := c.db.WithContext(ctx).Save(em); res.Error != nil {
		c.l.Warn("failed to update emitter status", zap.Error(res.Error))
	}

	return em, nil
}

func (c *Client) getQueue(ctx context.Context, queueID string) (*app.Queue, error) {
	var q app.Queue
	if res := c.db.WithContext(ctx).First(&q, "id = ?", queueID); res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get queue")
	}
	return &q, nil
}
