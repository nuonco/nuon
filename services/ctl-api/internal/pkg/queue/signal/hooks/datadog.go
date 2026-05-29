package hooks

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/datadogrender"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/hooks/slackrender"
)

// DatadogParams declares the dependencies for the DD signal lifecycle
// hook. All fields are optional (mirroring webhook.go's defensive
// constructor) so the hook can be wired into FX even when DD isn't
// configured for the env.
type DatadogParams struct {
	fx.In

	Cfg      *internal.Config `optional:"true"`
	L        *zap.Logger      `optional:"true"`
	DB       *gorm.DB         `name:"psql" optional:"true"`
	DDClient *ddclient.Client `optional:"true"`
	MW       metrics.Writer   `optional:"true"`
}

// DatadogSignalLifecycleHook fans out workflow / step / approval lifecycle
// events to all verified DatadogConnections for the event's org, evaluating
// per-row Match + Interests just like SlackChannelSubscriptions do.
//
// Unlike the Slack hook this has no threading and no per-channel dedup —
// DD's event stream is flat; we set aggregation_key = workflow_id so the
// DD UI groups events under their parent workflow on its own.
//
// Most enrichment (buildEventData, EventTargetsFromEvent, label loader) is
// shared with the webhook hook via a lightweight `enricher` delegate so we
// keep one source of truth for payload shape.
type DatadogSignalLifecycleHook struct {
	l        *zap.Logger
	db       *gorm.DB
	ddClient *ddclient.Client
	enricher *WebhookSignalLifecycleHook
	mw       metrics.Writer
}

var _ signal.SignalLifecycleHook = (*DatadogSignalLifecycleHook)(nil)

// NewDatadogSignalLifecycleHook constructs the DD lifecycle hook. Returns
// a non-nil hook even when dependencies are missing — Supports()
// short-circuits at runtime so the dispatcher stays cheap when DD isn't
// configured.
func NewDatadogSignalLifecycleHook(params DatadogParams) *DatadogSignalLifecycleHook {
	logger := params.L
	if logger == nil {
		logger = zap.NewNop()
	}

	appURL := ""
	publicAPIURL := ""
	if params.Cfg != nil {
		appURL = strings.TrimSpace(params.Cfg.AppURL)
		publicAPIURL = strings.TrimSpace(params.Cfg.PublicAPIURL)
	}

	// Reuse webhook.go's enrichment pipeline. Building a private instance
	// (rather than depending on the FX-wired hook) keeps DD and webhook
	// independently constructible and avoids accidental cycles in the
	// dependency graph. The enricher only ever reads from the DB.
	enricher := &WebhookSignalLifecycleHook{
		l:            logger,
		db:           params.DB,
		appURL:       appURL,
		publicAPIURL: publicAPIURL,
	}

	return &DatadogSignalLifecycleHook{
		l:        logger,
		db:       params.DB,
		ddClient: params.DDClient,
		enricher: enricher,
		mw:       params.MW,
	}
}

func (h *DatadogSignalLifecycleHook) Name() string {
	return "workflow_lifecycle_datadog"
}

// metricNamespace returns the Temporal namespace tag value for metrics
// emitted from inside an activity. Returns "" outside activity ctx.
func (h *DatadogSignalLifecycleHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

func (h *DatadogSignalLifecycleHook) emitPublishLatency(ctx context.Context, phasePrefix string, startTS time.Time) {
	if h.mw == nil {
		return
	}
	h.mw.Timing(
		fmt.Sprintf("signal_lifecycle.%s.datadog.publish_latency", phasePrefix),
		time.Since(startTS),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

func (h *DatadogSignalLifecycleHook) emitError(ctx context.Context, phasePrefix string) {
	if h.mw == nil {
		return
	}
	h.mw.Incr(
		fmt.Sprintf("signal_lifecycle.%s.datadog.errors", phasePrefix),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

// Supports limits this hook to the public lifecycle primitives (matches
// the webhook + slack hooks' filter) and short-circuits when DD isn't
// wired so the dispatcher doesn't pay the per-event cost.
func (h *DatadogSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if h.ddClient == nil || h.db == nil {
		return false
	}

	switch event.SignalType {
	case signalTypeExecuteWorkflow,
		signalTypeExecuteWorkflowStep,
		signalTypeWorkflowStepApprovalRequest,
		signalTypeWorkflowStepApprovalResponse,
		signalTypeDriftDetected:
		return true
	default:
		return false
	}
}

func (h *DatadogSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}
	// Same logic as slack/webhook: skip "started" emission for approval
	// signals (no meaningful "started" semantic) and for drift-detected
	// (its Execute is a no-op single-shot carrier).
	if isApprovalSignalType(event.SignalType) || event.SignalType == signalTypeDriftDetected {
		return signal.AllowPhaseDecision(), nil
	}

	if err := h.publish(ctx, event, nil); err != nil {
		h.l.Debug("failed to publish datadog lifecycle event",
			zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

func (h *DatadogSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}

	h.l.Debug("workflow lifecycle datadog after-phase",
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("phase", string(event.Phase)),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("status", string(outcome.Status)),
	)

	return h.publish(ctx, event, &outcome)
}

// publish iterates verified DatadogConnections for the event's org,
// evaluates each connection's subscriptions against the event's targets,
// and posts to DD for each (connection, subscription) match. Delivery
// errors are aggregated so a single failing tenant doesn't swallow
// others.
func (h *DatadogSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	phasePrefix := "before_phase"
	if outcome != nil {
		phasePrefix = "after_phase"
	}
	startTS := time.Now()
	delivered := false
	defer func() {
		if delivered {
			h.emitPublishLatency(ctx, phasePrefix, startTS)
		}
	}()

	if event.OrgID == "" || event.WorkflowID == "" {
		return nil
	}

	// Cheap indexed lookup BEFORE enrichment. Most orgs have no DD
	// integration so this short-circuits the dominant DB cost when no
	// connections are configured.
	var connections []app.DatadogConnection
	if err := h.db.WithContext(ctx).
		Where(app.DatadogConnection{
			OrgID:  event.OrgID,
			Status: app.DatadogConnectionStatusVerified,
		}).
		Find(&connections).Error; err != nil {
		h.emitError(ctx, phasePrefix)
		return fmt.Errorf("unable to list datadog connections for lifecycle: %w", err)
	}
	if len(connections) == 0 {
		return nil
	}

	// Reuse webhook.go's payload builder so the renderer sees the same
	// enriched shape webhook / slack consumers see (plus the same
	// hidden-step suppression rules).
	data, ok := h.enricher.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	rendered := buildRenderEvent(data)

	targets := EventTargetsFromEvent(ctx, h.db, event, data)
	labelLoader := newLabelLoader(h.db)

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("org_id", event.OrgID),
		zap.String("workflow_id", event.WorkflowID),
		zap.String("event_install_id", targets.InstallID),
		zap.String("event_component_id", targets.ComponentID),
		zap.String("event_action_id", targets.ActionID),
	)

	var sendErrs []error
	for _, conn := range connections {
		// Per-(connection, queue_signal_id) dedup: two subs on the same
		// connection both matching the same event only produce one post
		// per connection. Keyed by connection.ID since DD's
		// aggregation_key still groups under the workflow.
		seen := make(map[string]struct{})

		var subs []app.DatadogEventSubscription
		if err := h.db.WithContext(ctx).
			Where(app.DatadogEventSubscription{
				ConnectionID: conn.ID,
				OrgID:        event.OrgID,
			}).
			Find(&subs).Error; err != nil {
			logger.Warn("failed to list datadog subscriptions",
				zap.String("connection_id", conn.ID), zap.Error(err))
			sendErrs = append(sendErrs, err)
			h.emitError(ctx, phasePrefix)
			continue
		}

		for _, sub := range subs {
			if err := labelLoader.load(ctx, &targets); err != nil {
				logger.Warn("failed to load event labels",
					zap.Error(err))
				// Fail open: selector matches simply miss when label
				// set is empty.
			}

			if !sub.Match.Matches(targets) {
				continue
			}
			if !interests.Matches(event, outcome, h.db, sub.Interests) {
				continue
			}

			dedupKey := conn.ID + "|" + event.QueueSignalID
			if _, dup := seen[dedupKey]; dup {
				continue
			}
			seen[dedupKey] = struct{}{}

			if err := h.postToConnection(ctx, &conn, sub, rendered.event); err != nil {
				sendErrs = append(sendErrs, err)
				h.emitError(ctx, phasePrefix)
				logger.Warn("failed to deliver datadog lifecycle event",
					zap.String("connection_id", conn.ID),
					zap.String("subscription_id", sub.ID),
					zap.Error(err))

				// If DD rejected the API key, flip our connection
				// status so subsequent events skip this tenant.
				if isDatadogAuthError(err) {
					if mErr := h.markConnectionRevoked(ctx, conn.ID); mErr != nil {
						logger.Warn("failed to mark datadog connection revoked after auth failure",
							zap.String("connection_id", conn.ID),
							zap.Error(mErr))
					} else {
						logger.Info("marked datadog connection revoked after auth failure",
							zap.String("connection_id", conn.ID))
					}
					// Bail on remaining subs for this connection
					// — they share the same dead key.
					break
				}
				continue
			}
			delivered = true
		}
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}
	return nil
}

// postToConnection renders the DD event payload for the given subscription
// and posts it to the connection's tenant.
//
// The input event uses slackrender.Event because that's the renderer
// input contract shared with the slack hook — buildRenderEvent (in
// slack.go) already converts lifecycleEventData → slackrender.Event so
// both consumers see the same enriched shape.
func (h *DatadogSignalLifecycleHook) postToConnection(
	ctx context.Context,
	conn *app.DatadogConnection,
	sub app.DatadogEventSubscription,
	event slackrender.Event,
) error {
	extraTags := mergeExtraTags([]string(conn.DefaultTags), []string(sub.AdditionalTags))
	req := datadogrender.Build(event, extraTags, sub.AlertTypeOverride, sub.PriorityOverride)

	baseURL := ddclient.ResolveSiteURL(conn.Site)
	if _, err := h.ddClient.PostEvent(ctx, baseURL, conn.APIKey, req); err != nil {
		return fmt.Errorf("post datadog event: %w", err)
	}
	return nil
}

// mergeExtraTags concatenates two []string slices into a fresh slice.
// Dedup/sort happens inside datadogrender.Build so callers don't have to
// think about it.
func mergeExtraTags(a, b []string) []string {
	out := make([]string, 0, len(a)+len(b))
	out = append(out, a...)
	out = append(out, b...)
	return out
}

// markConnectionRevoked transitions a verified connection to revoked
// after DD rejected its credentials. Logged at INFO by the caller; the
// dashboard surfaces the revoked status so the user can re-enter keys.
func (h *DatadogSignalLifecycleHook) markConnectionRevoked(ctx context.Context, connectionID string) error {
	return h.db.WithContext(ctx).
		Model(&app.DatadogConnection{}).
		Where("id = ? AND status = ?", connectionID, app.DatadogConnectionStatusVerified).
		Update("status", app.DatadogConnectionStatusRevoked).
		Error
}

// isDatadogAuthError reports whether the error indicates the API key was
// rejected (HTTP 401 / 403). Network failures, 5xx, etc. don't qualify —
// those are transient and shouldn't flip our status.
func isDatadogAuthError(err error) bool {
	if err == nil {
		return false
	}
	var apiErr *ddclient.APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401 || apiErr.StatusCode == 403
	}
	return false
}
