package signal

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/temporal/temporalzap"
)

type SignalLifecycleActivitiesParams struct {
	fx.In

	Hooks []SignalLifecycleHook `group:"signal_lifecycle_hooks"`
}

type SignalLifecycleActivities struct {
	hooks []SignalLifecycleHook
}

func NewSignalLifecycleActivities(params SignalLifecycleActivitiesParams) *SignalLifecycleActivities {
	return &SignalLifecycleActivities{
		hooks: params.Hooks,
	}
}

type RunSignalLifecyclePreExecuteRequest struct {
	Event SignalPhaseEvent `json:"event" validate:"required"`
}

type RunSignalLifecyclePreExecuteResponse struct {
	Allow    bool           `json:"allow"`
	Reason   string         `json:"reason,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *SignalLifecycleActivities) RunSignalLifecyclePreExecute(ctx context.Context, req *RunSignalLifecyclePreExecuteRequest) (*RunSignalLifecyclePreExecuteResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("run signal lifecycle pre-execute request is nil")
	}

	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("queue_signal_id", req.Event.QueueSignalID),
		zap.String("queue_id", req.Event.QueueID),
		zap.String("signal_type", string(req.Event.SignalType)),
		zap.String("phase", string(req.Event.Phase)),
	)

	l.Info("running signal lifecycle pre-execute hooks", zap.Int("registered_hooks", len(a.hooks)))

	resp := &RunSignalLifecyclePreExecuteResponse{Allow: true}
	processedHooks := 0
	for _, hook := range a.hooks {
		if !hook.Supports(req.Event) {
			continue
		}

		processedHooks++
		hookStart := time.Now()
		decision, err := hook.PreExecute(ctx, req.Event)
		hookDur := time.Since(hookStart)
		if err != nil {
			l.Error("pre-execute hook failed", zap.String("hook", hook.Name()), zap.Duration("hook_duration", hookDur), zap.Error(err))
			return nil, fmt.Errorf("pre-execute hook %q failed: %w", hook.Name(), err)
		}

		l.Info("pre-execute hook completed", zap.String("hook", hook.Name()), zap.Duration("hook_duration", hookDur))

		if len(decision.Metadata) > 0 {
			resp.Metadata = mergeSignalLifecycleMetadata(resp.Metadata, decision.Metadata)
		}

		if !decision.Allow {
			resp.Allow = false
			resp.Reason = decision.Reason
			l.Warn("signal lifecycle phase blocked by hook",
				zap.String("hook", hook.Name()),
				zap.String("reason", decision.Reason))
			l.Info("completed signal lifecycle pre-execute hooks",
				zap.Int("hooks_processed", processedHooks),
				zap.Bool("allow", resp.Allow))
			return resp, nil
		}
	}

	l.Info("completed signal lifecycle pre-execute hooks",
		zap.Int("hooks_processed", processedHooks),
		zap.Bool("allow", resp.Allow))

	return resp, nil
}

type RunSignalLifecyclePostExecuteRequest struct {
	Event   SignalPhaseEvent   `json:"event" validate:"required"`
	Outcome SignalPhaseOutcome `json:"outcome" validate:"required"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *SignalLifecycleActivities) RunSignalLifecyclePostExecute(ctx context.Context, req *RunSignalLifecyclePostExecuteRequest) error {
	if req == nil {
		return fmt.Errorf("run signal lifecycle post-execute request is nil")
	}

	l := temporalzap.GetActivityLogger(ctx).With(
		zap.String("queue_signal_id", req.Event.QueueSignalID),
		zap.String("queue_id", req.Event.QueueID),
		zap.String("signal_type", string(req.Event.SignalType)),
		zap.String("phase", string(req.Event.Phase)),
		zap.String("status", string(req.Outcome.Status)),
	)

	l.Info("running signal lifecycle post-execute hooks", zap.Int("registered_hooks", len(a.hooks)))

	processedHooks := 0
	failedHooks := 0
	for _, hook := range a.hooks {
		if !hook.Supports(req.Event) {
			continue
		}

		processedHooks++
		hookStart := time.Now()
		if err := hook.PostExecute(ctx, req.Event, req.Outcome); err != nil {
			failedHooks++
			l.Error("post-execute hook failed",
				zap.String("hook", hook.Name()),
				zap.Duration("hook_duration", time.Since(hookStart)),
				zap.Error(err))
			continue
		}
		l.Info("post-execute hook completed", zap.String("hook", hook.Name()), zap.Duration("hook_duration", time.Since(hookStart)))
	}

	l.Info("completed signal lifecycle post-execute hooks",
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
