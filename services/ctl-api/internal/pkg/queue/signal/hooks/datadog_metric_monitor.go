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

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	ddclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/datadog/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// DatadogMetricMonitorParams declares the dependencies for the DD
// metric-mode managed-monitor hook. All optional so the hook can be
// wired in tests where DD isn't configured.
type DatadogMetricMonitorParams struct {
	fx.In

	Cfg      *internal.Config `optional:"true"`
	L        *zap.Logger      `optional:"true"`
	DB       *gorm.DB         `name:"psql" optional:"true"`
	DDClient *ddclient.Client `optional:"true"`
	MW       metrics.Writer   `optional:"true"`
}

// DatadogMetricMonitorHook fires `nuon.monitor.fired` counts into the DD
// connection backing each matching metric-mode managed monitor.
//
// The contract with the monitor-create path (see
// internal/app/datadog/service/monitor_presets.go::buildMetricMonitorQuery):
// each metric-mode managed monitor in DD queries the single tag
// `nuon_monitor_id:<row_id>`. This hook evaluates each lifecycle event
// against every metric-mode managed monitor in the event's org, and for
// each match submits one count point tagged with that monitor's row ID.
//
// Cardinality cap: only `nuon_monitor_id` is ever submitted as a metric
// tag. Total DD custom-metric series cap == count of metric-mode
// managed-monitor rows, regardless of how many installs / components /
// actions / labels feed into the matcher.
//
// Decoupling from event subscriptions: this hook doesn't depend on
// DatadogEventSubscription. An org with zero subscriptions but at least
// one metric-mode managed monitor still gets DD-side alerts because
// Nuon submits the firing metric directly. That's the whole point —
// metric mode lets the alerting path work without routing the full
// event stream into DD.
type DatadogMetricMonitorHook struct {
	l        *zap.Logger
	db       *gorm.DB
	ddClient *ddclient.Client
	enricher *WebhookSignalLifecycleHook
	mw       metrics.Writer
}

var _ signal.SignalLifecycleHook = (*DatadogMetricMonitorHook)(nil)

// NewDatadogMetricMonitorHook constructs the metric-mode hook. Returns a
// non-nil hook even when dependencies are missing — Supports() gates
// the fast path so unconfigured envs pay no per-event cost.
func NewDatadogMetricMonitorHook(params DatadogMetricMonitorParams) *DatadogMetricMonitorHook {
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

	enricher := &WebhookSignalLifecycleHook{
		l:            logger,
		db:           params.DB,
		appURL:       appURL,
		publicAPIURL: publicAPIURL,
	}

	return &DatadogMetricMonitorHook{
		l:        logger,
		db:       params.DB,
		ddClient: params.DDClient,
		enricher: enricher,
		mw:       params.MW,
	}
}

func (h *DatadogMetricMonitorHook) Name() string {
	return "workflow_lifecycle_datadog_metric_monitor"
}

func (h *DatadogMetricMonitorHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

func (h *DatadogMetricMonitorHook) emitSubmitLatency(ctx context.Context, phasePrefix string, startTS time.Time) {
	if h.mw == nil {
		return
	}
	h.mw.Timing(
		fmt.Sprintf("signal_lifecycle.%s.datadog_metric_monitor.submit_latency", phasePrefix),
		time.Since(startTS),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

func (h *DatadogMetricMonitorHook) emitError(ctx context.Context, phasePrefix string) {
	if h.mw == nil {
		return
	}
	h.mw.Incr(
		fmt.Sprintf("signal_lifecycle.%s.datadog_metric_monitor.errors", phasePrefix),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

// Supports mirrors the event-mode DD hook's vocabulary. Same lifecycle
// surface — only the routing differs.
func (h *DatadogMetricMonitorHook) Supports(event signal.SignalPhaseEvent) bool {
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

// BeforePhase is a no-op: the failure preset only matches AfterPhase
// outcomes (status:failed), and drift is a single-shot signal whose
// BeforePhase is bypassed for the same reason it's bypassed in the
// event-mode hook. Approval signals don't match either preset.
func (h *DatadogMetricMonitorHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	return signal.AllowPhaseDecision(), nil
}

func (h *DatadogMetricMonitorHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}
	return h.fire(ctx, event, &outcome)
}

// fire resolves the event into per-monitor matches and submits one
// `nuon.monitor.fired{nuon_monitor_id:<id>}` count per match into the
// DD connection backing each monitor. Failures are logged but never
// propagated — a transient DD outage must not break workflow
// lifecycle.
func (h *DatadogMetricMonitorHook) fire(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	phasePrefix := "after_phase"
	startTS := time.Now()
	delivered := false
	defer func() {
		if delivered {
			h.emitSubmitLatency(ctx, phasePrefix, startTS)
		}
	}()

	if event.OrgID == "" || event.WorkflowID == "" {
		return nil
	}

	// Cheap indexed scan before enrichment: most orgs won't have
	// metric-mode monitors, and even orgs that do typically have
	// a small handful. JOIN connection so we have site/apiKey
	// available for the metric submission without a second query.
	var monitors []app.DatadogManagedMonitor
	if err := h.db.WithContext(ctx).
		Preload("Connection").
		Where(app.DatadogManagedMonitor{
			OrgID:  event.OrgID,
			Mode:   app.DatadogManagedMonitorModeMetric,
			Status: app.DatadogManagedMonitorStatusActive,
		}).
		Find(&monitors).Error; err != nil {
		h.emitError(ctx, phasePrefix)
		return fmt.Errorf("unable to list datadog metric monitors: %w", err)
	}
	if len(monitors) == 0 {
		return nil
	}

	// Enrichment is shared with the other lifecycle hooks. We need
	// data.Outcome.Status / data.Kind for preset matching and
	// EventTargetsFromEvent for the install/component/action ids.
	data, ok := h.enricher.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}
	targets := EventTargetsFromEvent(ctx, h.db, event, data)

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("org_id", event.OrgID),
		zap.String("workflow_id", event.WorkflowID),
	)

	// Group matches by (connection.ID) so we submit one PostSeries
	// per DD tenant rather than one HTTP request per monitor. Inside
	// the request the points carry distinct `nuon_monitor_id` tags
	// so DD demuxes them server-side.
	type batch struct {
		baseURL string
		apiKey  string
		series  []ddclient.MetricSeries
	}
	batches := make(map[string]*batch)

	now := time.Now().UTC().Unix()

	for i := range monitors {
		m := &monitors[i]
		if !monitorMatchesEvent(m, event, data, targets) {
			continue
		}
		conn := m.Connection
		if conn.Status != app.DatadogConnectionStatusVerified || conn.APIKey == "" {
			// Skip silently: a revoked connection emits user-
			// facing signal via the dashboard. Submitting against
			// a known-bad key would just generate noise in our
			// own logs.
			continue
		}

		b, ok := batches[conn.ID]
		if !ok {
			b = &batch{
				baseURL: ddclient.ResolveSiteURL(conn.Site),
				apiKey:  conn.APIKey,
			}
			batches[conn.ID] = b
		}
		b.series = append(b.series, ddclient.MetricSeries{
			Metric: ddclient.MonitorFiredMetric,
			Type:   ddclient.MetricTypeCount,
			Tags:   []string{"nuon_monitor_id:" + m.ID},
			Points: []ddclient.MetricPoint{{Timestamp: now, Value: 1}},
		})
	}

	if len(batches) == 0 {
		return nil
	}

	var sendErrs []error
	for connID, b := range batches {
		if err := h.ddClient.PostSeries(ctx, b.baseURL, b.apiKey, b.series); err != nil {
			h.emitError(ctx, phasePrefix)
			sendErrs = append(sendErrs, err)
			logger.Warn("failed to submit nuon.monitor.fired series to datadog",
				zap.String("connection_id", connID),
				zap.Int("series_count", len(b.series)),
				zap.Error(err))
			continue
		}
		delivered = true
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}
	return nil
}

// monitorMatchesEvent reproduces the predicate buildMonitorQuery encodes
// for event-mode managed monitors but evaluates it on Nuon's side so
// the only thing crossing the wire to DD is the bare metric tagged with
// the monitor's row ID. Keeping the two predicates aligned is the whole
// safety story for metric mode — if this drifts from buildMonitorQuery,
// users with both monitor modes side-by-side will see inconsistent
// alerting on identical events.
func monitorMatchesEvent(
	m *app.DatadogManagedMonitor,
	event signal.SignalPhaseEvent,
	data lifecycleEventData,
	targets labels.EventTargets,
) bool {
	if !targetMatches(m, event, targets) {
		return false
	}
	// Optional install scope (mirrors buildMonitorQuery's
	// `nuon_install_id:<install>` AND clause for non-install targets).
	if m.InstallID != "" && m.TargetType != app.DatadogManagedMonitorTargetTypeInstall {
		if targets.InstallID != m.InstallID {
			return false
		}
	}
	return presetMatches(m.Preset, event, data)
}

// targetMatches mirrors the per-target-type selector in buildMonitorQuery.
// Each branch maps to one of the `nuon_<target>_id:<id>` tags the
// renderer emits.
func targetMatches(
	m *app.DatadogManagedMonitor,
	event signal.SignalPhaseEvent,
	targets labels.EventTargets,
) bool {
	switch m.TargetType {
	case app.DatadogManagedMonitorTargetTypeInstall:
		return targets.InstallID == m.TargetID
	case app.DatadogManagedMonitorTargetTypeComponent:
		return targets.ComponentID == m.TargetID
	case app.DatadogManagedMonitorTargetTypeWorkflow:
		return event.WorkflowID == m.TargetID
	case app.DatadogManagedMonitorTargetTypeAction:
		return targets.ActionID == m.TargetID
	default:
		return false
	}
}

// presetMatches mirrors the conditionTag clause in buildMonitorQuery.
//
//   - failure  → nuon_status:failed (outcome reported as "failed")
//   - drift    → drift-detected signal type (the renderer doesn't actually
//     emit nuon_kind:drift in v1; the metric path keys off
//     the raw signal type directly so this preset behaves
//     correctly regardless of the event-mode tag drift).
func presetMatches(preset app.DatadogManagedMonitorPreset, event signal.SignalPhaseEvent, data lifecycleEventData) bool {
	switch preset {
	case app.DatadogManagedMonitorPresetFailure:
		return data.Outcome != nil && data.Outcome.Status == statusFailed
	case app.DatadogManagedMonitorPresetDrift:
		return event.SignalType == signalTypeDriftDetected
	default:
		return false
	}
}
