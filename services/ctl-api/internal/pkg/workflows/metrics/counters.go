package workflowmetrics

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type CounterParams struct {
	fx.In

	Config   *telemetry.Config
	Provider metric.MeterProvider
	DB       *gorm.DB `name:"psql"`
	L        *zap.Logger
}

type Counters struct {
	completed   metric.Int64Counter
	retries     metric.Int64Counter
	startDelay  metric.Float64Histogram
	elapsedTime metric.Float64Histogram
	db          *gorm.DB
	l           *zap.Logger
}

func NewCounters(params CounterParams) (*Counters, error) {
	c := &Counters{}
	if params.Config.Endpoint == "" {
		return c, nil
	}
	c.db, c.l = params.DB, params.L
	meter := params.Provider.Meter("github.com/nuonco/nuon/services/ctl-api/workflows")
	var err error
	c.completed, err = meter.Int64Counter("nuon.workflow.executions.completed", metric.WithUnit("{execution}"),
		metric.WithDescription("Completed execute-workflow queue-signal executions, classified by persisted workflow outcome."))
	if err != nil {
		return nil, err
	}
	c.retries, err = meter.Int64Counter("nuon.workflow.step.retries", metric.WithUnit("{retry}"),
		metric.WithDescription("Persisted step retry decisions, not attempts started or Temporal activity retries."))
	if err != nil {
		return nil, err
	}
	c.startDelay, err = meter.Float64Histogram("nuon.workflow.start_delay", metric.WithUnit("s"),
		metric.WithDescription("Time from workflow creation to its first persisted execution start."),
		metric.WithExplicitBucketBoundaries(1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600))
	if err != nil {
		return nil, err
	}
	c.elapsedTime, err = meter.Float64Histogram("nuon.workflow.elapsed_time", metric.WithUnit("s"),
		metric.WithDescription("Time from workflow creation to observed execution completion, including queueing, approval and retry waits."),
		metric.WithExplicitBucketBoundaries(10, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200, 21600, 86400))
	if err != nil {
		return nil, err
	}
	for _, typ := range workflowTypes {
		for _, outcome := range []string{"success", "error", "cancelled", "unknown"} {
			c.completed.Add(context.Background(), 0, metric.WithAttributes(
				attribute.String("workflow.type", typ), attribute.String("workflow.outcome", outcome)))
		}
		for _, source := range []string{"auto", "manual"} {
			c.retries.Add(context.Background(), 0, metric.WithAttributes(
				attribute.String("workflow.type", typ), attribute.String("retry.source", source)))
		}
	}
	return c, nil
}

func terminal(status app.Status) bool {
	return status == app.StatusSuccess || status == app.StatusError || status == app.StatusCancelled
}

func executionCompleted(before app.QueueSignal, after app.Status) bool {
	return before.Type == "execute-workflow" && before.OwnerType == "install_workflows" && before.OwnerID != "" &&
		before.DeletedAt == 0 && !terminal(before.Status.Status) && terminal(after)
}

func executionOutcome(status app.Status) string {
	if terminal(status) {
		return string(status)
	}
	return "unknown"
}

func (c *Counters) SignalStatusUpdated(ctx context.Context, before app.QueueSignal, after app.Status) {
	if c == nil || c.completed == nil || !executionCompleted(before, after) {
		return
	}
	completedAt := time.Now()
	wf, ok := c.workflow(ctx, before.OwnerID)
	if !ok {
		return
	}
	attrs := metric.WithAttributes(attribute.String("workflow.type", string(wf.Type)),
		attribute.String("workflow.outcome", executionOutcome(wf.Status.Status)))
	c.completed.Add(ctx, 1, attrs)
	if seconds, ok := elapsedSeconds(wf.CreatedAt, completedAt); ok {
		c.elapsedTime.Record(ctx, seconds, attrs)
	}
}

func elapsedSeconds(start, end time.Time) (float64, bool) {
	return end.Sub(start).Seconds(), !start.IsZero() && !end.IsZero() && !end.Before(start)
}

func (c *Counters) FlowStarted(ctx context.Context, started app.Workflow) {
	if c == nil || c.startDelay == nil {
		return
	}
	wf, ok := c.workflow(ctx, started.ID)
	if !ok {
		return
	}
	if seconds, ok := elapsedSeconds(wf.CreatedAt, started.StartedAt); ok {
		c.startDelay.Record(ctx, seconds, metric.WithAttributes(attribute.String("workflow.type", string(wf.Type))))
	}
}

func retrySource(status app.CompositeStatus) string {
	if status.Status == app.StatusError && status.Metadata["auto_retried"] == true {
		return "auto"
	}
	if status.Status == app.StatusDiscarded && status.Metadata["retry_type"] == "manual" {
		return "manual"
	}
	return ""
}

func newRetryDecision(before app.CompositeStatus, source string) bool {
	switch source {
	case "auto":
		return before.Metadata["auto_retried"] != true
	case "manual":
		return before.Status != app.StatusDiscarded && before.Metadata["retry_type"] != "manual"
	default:
		return false
	}
}

func (c *Counters) StepStatusUpdated(ctx context.Context, before app.WorkflowStep, requested app.CompositeStatus) {
	source := retrySource(requested)
	if c == nil || c.retries == nil || before.DeletedAt != 0 || !newRetryDecision(before.Status, source) {
		return
	}
	wf, ok := c.workflow(ctx, before.InstallWorkflowID)
	if !ok {
		return
	}
	c.retries.Add(ctx, 1, metric.WithAttributes(
		attribute.String("workflow.type", string(wf.Type)), attribute.String("retry.source", source)))
}

func (c *Counters) workflow(ctx context.Context, id string) (app.Workflow, bool) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	var wf app.Workflow
	err := c.db.WithContext(ctx).
		Select("install_workflows.id", "install_workflows.type", "install_workflows.plan_only", "install_workflows.status", "install_workflows.created_at").
		Joins("JOIN installs ON installs.id = install_workflows.owner_id AND installs.deleted_at = 0").
		Where(app.Workflow{ID: id, OwnerType: "installs"}).
		Take(&wf).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			c.l.Warn("unable to load workflow for telemetry", zap.Error(err))
		}
		return wf, false
	}
	return wf, !wf.PlanOnly && slices.Contains(workflowTypes, string(wf.Type))
}
