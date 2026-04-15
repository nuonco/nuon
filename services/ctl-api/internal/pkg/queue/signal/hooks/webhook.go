package hooks

import (
	"context"
	"net/http"
	"time"

	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/webhook"
)

// userFacingOperations is the set of signal operations that should be exposed
// via webhooks. Internal-only signals are filtered out.
var userFacingOperations = map[string]struct{}{
	"component-deploy":    {},
	"component-teardown":  {},
	"sandbox-provision":   {},
	"sandbox-deprovision": {},
	"sandbox-reprovision": {},
	"runner-provision":    {},
	"runner-reprovision":  {},
	"action-workflow-run": {},
	"install-created":     {},
	"install-updated":     {},
	"install-restart":     {},
}

type Params struct {
	fx.In

	Cfg *internal.Config `optional:"true"`
	L   *zap.Logger      `optional:"true"`
	DB  *gorm.DB         `name:"psql" optional:"true"`
}

type WebhookSignalLifecycleHook struct {
	l         *zap.Logger
	deliverer *webhook.Deliverer
}

var _ signal.SignalLifecycleHook = (*WebhookSignalLifecycleHook)(nil)

func NewWebhookSignalLifecycleHook(params Params) *WebhookSignalLifecycleHook {
	logger := params.L
	if logger == nil {
		logger = zap.NewNop()
	}

	timeout := 5 * time.Second
	if params.Cfg != nil && params.Cfg.WebhookTimeout > 0 {
		timeout = params.Cfg.WebhookTimeout
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	webhookURLs := []string{}
	if params.Cfg != nil {
		webhookURLs = params.Cfg.WebhookURLs
	}

	return &WebhookSignalLifecycleHook{
		l: logger,
		deliverer: webhook.NewDeliverer(
			logger,
			&http.Client{Timeout: timeout},
			webhookURLs,
			params.DB,
		),
	}
}

func (h *WebhookSignalLifecycleHook) Name() string {
	return "signal_lifecycle_webhook"
}

func (h *WebhookSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if h.deliverer == nil {
		return false
	}

	if event.Operation == "" {
		h.l.Debug("webhook hook skipped: empty operation",
			zap.String("queue_signal_id", event.QueueSignalID),
			zap.String("phase", string(event.Phase)),
		)
		return false
	}

	if event.Phase == signal.SignalPhaseValidate {
		return false
	}

	_, ok := userFacingOperations[event.Operation]
	return ok
}

func (h *WebhookSignalLifecycleHook) PreExecute(ctx context.Context, event signal.SignalPhaseEvent) (signal.PreExecuteDecision, error) {
	if err := h.deliverer.Publish(ctx, event, nil, "before"); err != nil {
		h.l.Debug("failed to publish pre-execute signal lifecycle webhook", zap.Error(err))
	}

	return signal.AllowDecision(), nil
}

func (h *WebhookSignalLifecycleHook) PostExecute(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	h.l.Debug("webhook after-phase called",
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("phase", string(event.Phase)),
		zap.String("operation", event.Operation),
		zap.String("status", string(outcome.Status)),
	)
	return h.deliverer.Publish(ctx, event, &outcome, "after")
}
