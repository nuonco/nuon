package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// nuonEventsTable is the ClickHouse table written by this hook. Kept as
// a constant rather than a GORM model + AutoMigrate so the SQL schema in
// internal/pkg/db/ch/migrations/09_create_nuon_events_table.sql remains
// the single source of truth — the row struct below is just a transport
// shape and only its column tags must agree with the migration.
const nuonEventsTable = "nuon_events"

// nuonEventRow mirrors the columns declared in
// internal/pkg/db/ch/migrations/09_create_nuon_events_table.sql.
// Column tags are explicit so a future struct field rename can't silently
// drift away from the schema.
type nuonEventRow struct {
	EventID    string    `gorm:"column:event_id"`
	Timestamp  time.Time `gorm:"column:ts"`
	SignalType string    `gorm:"column:signal_type"`
	Phase      string    `gorm:"column:phase"`
	Transition string    `gorm:"column:transition"`
	Kind       string    `gorm:"column:kind"`

	OrgID string `gorm:"column:org_id"`

	InstallID    string `gorm:"column:install_id"`
	ComponentID  string `gorm:"column:component_id"`
	WorkflowID   string `gorm:"column:workflow_id"`
	ActionID     string `gorm:"column:action_id"`
	WorkflowType string `gorm:"column:workflow_type"`

	Status       string `gorm:"column:status"`
	OutcomeError string `gorm:"column:outcome_error"`
	DurationMs   int64  `gorm:"column:duration_ms"`
	ApprovalType string `gorm:"column:approval_type"`

	Tags    []string `gorm:"column:tags;type:Array(String)"`
	Payload string   `gorm:"column:payload"`
}

// ClickHouseParams declares the dependencies for the CH signal lifecycle
// sink. All fields are optional so the hook can be wired into FX even
// when CH or features aren't configured (e.g. in tests that don't spin
// the analytics stack).
type ClickHouseParams struct {
	fx.In

	Cfg      *internal.Config   `optional:"true"`
	L        *zap.Logger        `optional:"true"`
	DB       *gorm.DB           `name:"psql" optional:"true"`
	CHDB     *gorm.DB           `name:"ch" optional:"true"`
	Features *features.Features `optional:"true"`
	MW       metrics.Writer     `optional:"true"`
}

// ClickHouseSignalLifecycleHook mirrors every supported signal lifecycle
// event into the per-org `nuon_events` ClickHouse table.
//
// Two consumers depend on this sink:
//
//  1. The dashboard event feed reads from `nuon_events` so users can
//     browse the same lifecycle stream that webhooks / Slack / DD see,
//     in deployments where no outbound integration is configured.
//
//  2. The metric-mode Datadog managed monitors (next commit) read this
//     sink to evaluate match / interests on the Nuon side and then
//     submit a single `nuon.monitor.fired{nuon_monitor_id:<id>}` metric
//     to DD — without the CH sink in place the metric-mode path would
//     need to re-derive lifecycle outcomes from Postgres on every tick.
//
// The hook is gated by the per-org feature flag
// app.OrgFeatureClickHouseEvents — a tenant that hasn't opted into the
// durable event log pays no write cost. The feature flag is also the
// gate the metric-mode monitor path checks before claiming it can fire
// without a DD event subscription, so both consumers see a consistent
// view of "this org has a CH event stream".
//
// Enrichment (buildEventData, EventTargetsFromEvent) is shared with
// webhook / Slack / DD via the same lightweight enricher delegate the
// DD hook uses — one source of truth for payload shape across every
// sink the platform exposes.
type ClickHouseSignalLifecycleHook struct {
	l        *zap.Logger
	chDB     *gorm.DB
	features *features.Features
	enricher *WebhookSignalLifecycleHook
	mw       metrics.Writer
}

var _ signal.SignalLifecycleHook = (*ClickHouseSignalLifecycleHook)(nil)

// NewClickHouseSignalLifecycleHook constructs the CH lifecycle sink.
// Returns a non-nil hook even when dependencies are missing — Supports()
// short-circuits at runtime so the dispatcher stays cheap when CH isn't
// wired (e.g. in unit tests for upstream packages).
func NewClickHouseSignalLifecycleHook(params ClickHouseParams) *ClickHouseSignalLifecycleHook {
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

	// Reuse webhook.go's enrichment pipeline. Mirrors the DD hook: an
	// inline enricher (rather than a separate FX-wired hook) keeps the
	// dependency graph acyclic and the sinks independently constructible.
	// The enricher only ever reads from the primary Postgres DB; the CH
	// db handle here is exclusively for writes into `nuon_events`.
	enricher := &WebhookSignalLifecycleHook{
		l:            logger,
		db:           params.DB,
		appURL:       appURL,
		publicAPIURL: publicAPIURL,
	}

	return &ClickHouseSignalLifecycleHook{
		l:        logger,
		chDB:     params.CHDB,
		features: params.Features,
		enricher: enricher,
		mw:       params.MW,
	}
}

func (h *ClickHouseSignalLifecycleHook) Name() string {
	return "workflow_lifecycle_clickhouse"
}

// metricNamespace mirrors the helper on the DD / webhook hooks so emitted
// timing tags identify which worker namespace produced the write.
func (h *ClickHouseSignalLifecycleHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

func (h *ClickHouseSignalLifecycleHook) emitWriteLatency(ctx context.Context, phasePrefix string, startTS time.Time) {
	if h.mw == nil {
		return
	}
	h.mw.Timing(
		fmt.Sprintf("signal_lifecycle.%s.clickhouse.write_latency", phasePrefix),
		time.Since(startTS),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

func (h *ClickHouseSignalLifecycleHook) emitError(ctx context.Context, phasePrefix string) {
	if h.mw == nil {
		return
	}
	h.mw.Incr(
		fmt.Sprintf("signal_lifecycle.%s.clickhouse.errors", phasePrefix),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

// Supports gates the dispatcher fast path. Same vocabulary as the other
// lifecycle hooks (workflow / step / approval / drift) so the CH stream
// is the union of everything the public sinks see.
func (h *ClickHouseSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if h.chDB == nil || h.features == nil || h.enricher == nil || h.enricher.db == nil {
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

func (h *ClickHouseSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}
	// Mirror the slack / webhook / DD hooks: skip "started" emission for
	// approval signals (no meaningful "started" semantic) and for
	// drift-detected (its Execute is a no-op single-shot carrier).
	if isApprovalSignalType(event.SignalType) || event.SignalType == signalTypeDriftDetected {
		return signal.AllowPhaseDecision(), nil
	}

	if err := h.write(ctx, event, nil); err != nil {
		h.l.Debug("failed to write clickhouse lifecycle event",
			zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

func (h *ClickHouseSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}
	return h.write(ctx, event, &outcome)
}

// write builds the normalized lifecycle row and inserts it into
// `nuon_events` for the event's org, gated by the per-org feature flag.
// Failures are logged but do not propagate — the CH sink is a
// best-effort secondary write; dropping it must never break a workflow.
func (h *ClickHouseSignalLifecycleHook) write(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	phasePrefix := "before_phase"
	if outcome != nil {
		phasePrefix = "after_phase"
	}
	startTS := time.Now()

	if event.OrgID == "" || event.WorkflowID == "" {
		return nil
	}

	// Org feature gate. Skipped tenants pay only this single SELECT on
	// orgs.features (already cached in PG buffers) — no enrichment, no
	// CH round-trip. Failing to read the flag fails closed (skip the
	// write) so a transient features service hiccup never floods CH
	// with rows from orgs that haven't opted in.
	enabled, err := h.features.OrgHasFeature(ctx, event.OrgID, app.OrgFeatureClickHouseEvents)
	if err != nil {
		h.l.Debug("failed to check clickhouse-events feature for org",
			zap.String("org_id", event.OrgID),
			zap.Error(err))
		return nil
	}
	if !enabled {
		return nil
	}

	// Reuse the shared enricher so CH rows see the same enriched shape
	// the other sinks emit (and the same hidden-step suppression rules).
	data, ok := h.enricher.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	targets := EventTargetsFromEvent(ctx, h.enricher.db, event, data)

	payload, err := json.Marshal(data)
	if err != nil {
		h.l.Debug("failed to marshal clickhouse lifecycle payload",
			zap.Error(err))
		return nil
	}

	row := nuonEventRow{
		EventID:      uuid.New().String(),
		Timestamp:    time.Now().UTC(),
		SignalType:   string(event.SignalType),
		Phase:        string(event.Phase),
		Transition:   data.Transition,
		Kind:         data.Kind,
		OrgID:        event.OrgID,
		InstallID:    targets.InstallID,
		ComponentID:  targets.ComponentID,
		WorkflowID:   event.WorkflowID,
		ActionID:     targets.ActionID,
		WorkflowType: event.WorkflowType,
		Tags:         interests.Classify(event, outcome, h.enricher.db),
		Payload:      string(payload),
	}
	if data.Outcome != nil {
		row.Status = data.Outcome.Status
		row.OutcomeError = data.Outcome.Error
		row.DurationMs = data.Outcome.DurationMs
	}
	if data.Approval != nil {
		row.ApprovalType = data.Approval.Type
	}

	if err := h.chDB.WithContext(ctx).
		Table(nuonEventsTable).
		Create(&row).Error; err != nil {
		h.emitError(ctx, phasePrefix)
		h.l.Warn("failed to insert nuon_events row",
			zap.String("org_id", event.OrgID),
			zap.String("workflow_id", event.WorkflowID),
			zap.String("signal_type", string(event.SignalType)),
			zap.Error(err))
		return nil
	}

	h.emitWriteLatency(ctx, phasePrefix, startTS)
	return nil
}
