package signal

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

type SignalLifecycleActivitiesParams struct {
	fx.In

	Hooks         []SignalLifecycleHook `group:"signal_lifecycle_hooks"`
	MW            metrics.Writer        `optional:"true"`
	MeterProvider metric.MeterProvider  `optional:"true"`
}

type SignalLifecycleActivities struct {
	hooks           []SignalLifecycleHook
	mw              metrics.Writer
	hookInvocations metric.Int64Counter
}

func NewSignalLifecycleActivities(params SignalLifecycleActivitiesParams) *SignalLifecycleActivities {
	a := &SignalLifecycleActivities{
		hooks: params.Hooks,
		mw:    params.MW,
	}
	if params.MeterProvider != nil {
		a.hookInvocations, _ = params.MeterProvider.Meter("github.com/nuonco/nuon/services/ctl-api/event-hooks").Int64Counter(
			"nuon.event.hook.invocations",
			metric.WithUnit("{invocation}"),
			metric.WithDescription("Returned signal lifecycle hook invocation outcomes."),
		)
	}
	return a
}

func (a *SignalLifecycleActivities) recordHookInvocation(ctx context.Context, hook SignalLifecycleHook, event SignalPhaseEvent, invocation, outcome string) {
	if a.hookInvocations == nil {
		return
	}
	a.hookInvocations.Add(ctx, 1, metric.WithAttributes(
		attribute.String("nuon.event.hook.name", boundedHookName(hook.Name())),
		attribute.String("nuon.event.hook.phase", boundedHookPhase(event.Phase)),
		attribute.String("nuon.event.hook.invocation", invocation),
		attribute.String("nuon.event.hook.outcome", outcome),
	))
}

func boundedHookName(name string) string {
	switch name {
	case "flow_lifecycle_telemetry", "workflow_lifecycle_webhook", "workflow_lifecycle_slack":
		return name
	default:
		return "other"
	}
}

func boundedHookPhase(phase SignalPhase) string {
	switch phase {
	case SignalPhaseValidate, SignalPhaseExecute, SignalPhaseCancel:
		return string(phase)
	default:
		return "other"
	}
}

// emitActivityLatency records the wrapper activity's wall-clock duration
// tagged by Temporal namespace. Only emits when a metrics writer is wired
// (tests skip metrics setup).
func (a *SignalLifecycleActivities) emitActivityLatency(ctx context.Context, metricName string, startTS time.Time) {
	if a.mw == nil {
		return
	}
	namespace := ""
	if info := activity.GetInfo(ctx); info.WorkflowNamespace != "" {
		namespace = info.WorkflowNamespace
	}
	a.mw.Timing(metricName, time.Since(startTS), metrics.ToTags(map[string]string{
		"namespace": namespace,
	}))
}

type RunSignalLifecycleBeforePhaseRequest struct {
	Event SignalPhaseEvent `json:"event" validate:"required"`
}

type RunSignalLifecycleBeforePhaseResponse struct {
	Allow    bool           `json:"allow"`
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *SignalLifecycleActivities) RunSignalLifecycleBeforePhase(ctx context.Context, req *RunSignalLifecycleBeforePhaseRequest) (*RunSignalLifecycleBeforePhaseResponse, error) {
	startTS := time.Now()
	defer a.emitActivityLatency(ctx, "signal_lifecycle.before_phase.activity_latency", startTS)

	if req == nil {
		return nil, fmt.Errorf("run signal lifecycle before-phase request is nil")
	}

	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("queue_signal_id", req.Event.QueueSignalID),
		zap.String("queue_id", req.Event.QueueID),
		zap.String("signal_type", string(req.Event.SignalType)),
		zap.String("phase", string(req.Event.Phase)),
	)

	l.Debug("running signal lifecycle before-phase hooks", zap.Int("registered_hooks", len(a.hooks)))

	resp := &RunSignalLifecycleBeforePhaseResponse{Allow: true}
	processedHooks := 0
	for _, hook := range a.hooks {
		if !hook.Supports(req.Event) {
			continue
		}

		processedHooks++
		decision, err := hook.BeforePhase(ctx, req.Event)
		if err != nil {
			a.recordHookInvocation(ctx, hook, req.Event, "before", "error")
			l.Error("before-phase hook failed", zap.String("hook", hook.Name()), zap.Error(err))
			return nil, fmt.Errorf("before-phase hook %q failed: %w", hook.Name(), err)
		}

		if len(decision.Metadata) > 0 {
			resp.Metadata = mergeSignalLifecycleMetadata(resp.Metadata, decision.Metadata)
		}

		if !decision.Allow {
			a.recordHookInvocation(ctx, hook, req.Event, "before", "blocked")
			resp.Allow = false
			resp.Reason = decision.Reason
			l.Warn("signal lifecycle phase blocked by hook",
				zap.String("hook", hook.Name()),
				zap.String("reason", decision.Reason))
			l.Debug("completed signal lifecycle before-phase hooks",
				zap.Int("hooks_processed", processedHooks),
				zap.Bool("allow", resp.Allow))
			return resp, nil
		}
		a.recordHookInvocation(ctx, hook, req.Event, "before", "success")
	}

	l.Debug("completed signal lifecycle before-phase hooks",
		zap.Int("hooks_processed", processedHooks),
		zap.Bool("allow", resp.Allow))

	return resp, nil
}

type RunSignalLifecycleAfterPhaseRequest struct {
	Event   SignalPhaseEvent   `json:"event" validate:"required"`
	Outcome SignalPhaseOutcome `json:"outcome" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *SignalLifecycleActivities) RunSignalLifecycleAfterPhase(ctx context.Context, req *RunSignalLifecycleAfterPhaseRequest) error {
	startTS := time.Now()
	defer a.emitActivityLatency(ctx, "signal_lifecycle.after_phase.activity_latency", startTS)

	if req == nil {
		return fmt.Errorf("run signal lifecycle after-phase request is nil")
	}

	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("queue_signal_id", req.Event.QueueSignalID),
		zap.String("queue_id", req.Event.QueueID),
		zap.String("signal_type", string(req.Event.SignalType)),
		zap.String("phase", string(req.Event.Phase)),
		zap.String("status", string(req.Outcome.Status)),
	)

	l.Debug("running signal lifecycle after-phase hooks", zap.Int("registered_hooks", len(a.hooks)))

	processedHooks := 0
	failedHooks := 0
	for _, hook := range a.hooks {
		if !hook.Supports(req.Event) {
			continue
		}

		processedHooks++
		if err := hook.AfterPhase(ctx, req.Event, req.Outcome); err != nil {
			a.recordHookInvocation(ctx, hook, req.Event, "after", "error")
			failedHooks++
			l.Error("after-phase hook failed",
				zap.String("hook", hook.Name()),
				zap.Error(err))
		} else {
			a.recordHookInvocation(ctx, hook, req.Event, "after", "success")
		}
	}

	l.Debug("completed signal lifecycle after-phase hooks",
		zap.Int("hooks_processed", processedHooks),
		zap.Int("hooks_failed", failedHooks))

	return nil
}

func mergeSignalLifecycleMetadata(dst, src map[string]any) map[string]any {
	if len(src) == 0 {
		return dst
	}

	if dst == nil {
		dst = make(map[string]any, len(src))
	}

	for k, v := range src {
		dst[k] = v
	}

	return dst
}
