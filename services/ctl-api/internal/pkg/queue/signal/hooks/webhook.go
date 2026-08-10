package hooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Public webhook primitives. Consumers reason about three things: the workflow
// lifecycle, the workflow step lifecycle, and the approval handshake on a
// step. Operation taxonomy (component-deploy, sandbox-provision, etc.),
// multi-phase concepts (plan/apply), and inner signal type names are
// deliberately NOT exposed.
const (
	cloudEventTypeWorkflow                  = "com.nuon.workflow.lifecycle.v1"
	cloudEventTypeWorkflowStep              = "com.nuon.workflow_step.lifecycle.v1"
	cloudEventTypeWorkflowStepApproval      = "com.nuon.workflow_step.approval.v1"
	cloudEventTypeWorkflowStepAwaitingRetry = "com.nuon.workflow_step.awaiting_retry.v1"
	cloudEventTypeStackRun                  = "com.nuon.stack.run.v1"
	cloudEventTypeRoleChange                = "com.nuon.stack.role_change.v1"
	cloudEventTypeInputsUpdated             = "com.nuon.stack.inputs_updated.v1"
	cloudEventTypeAppConfigSynced           = "com.nuon.app.config_synced.v1"
	cloudEventTypeUpdateAppConfig           = "com.nuon.install.app_config_updated.v1"
	cloudEventTypeRunnerUnhealthy           = "com.nuon.runner.unhealthy.v1"
	// Component health emits one CloudEvent type per level; `transition`
	// distinguishes going bad from recovering, the way the approval event uses
	// requested / approved / rejected.
	cloudEventTypeComponentHealth   = "com.nuon.component.health.v1"
	cloudEventTypeInstallHealth     = "com.nuon.install.health.v1"
	cloudEventTypeInstallSync       = "com.nuon.app.install_sync.v1"
	cloudEventTypeInstallConfigSync = "com.nuon.install.config_sync.v1"

	kindWorkflow             = "workflow"
	kindWorkflowStep         = "workflow_step"
	kindWorkflowStepApproval = "workflow_step_approval"
	kindStackRun             = "stack_run"
	kindRoleChange           = "role_change"
	kindInputsUpdated        = "inputs_updated"
	kindAppConfigSynced      = "app_config_synced"
	kindUpdateAppConfig      = "app_config_updated"
	kindRunnerUnhealthy      = "runner_unhealthy"
	kindComponentHealth      = "component_health"
	kindInstallHealth        = "install_health"
	kindInstallSync          = "install_sync"
	kindInstallConfigSync    = "install_config_sync"
)

// Status values surfaced to webhook consumers in the *.lifecycle events.
const (
	statusStarted   = "started"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
	statusCanceled  = "cancelled"
)

// Transition values surfaced in `data.transition`. Workflow / step events
// emit started / succeeded / failed / cancelled. Approval events emit a
// distinct vocabulary — requested when the approval row is created and
// approved / rejected when a response lands.
const (
	transitionStarted   = "started"
	transitionSucceeded = "succeeded"
	transitionFailed    = "failed"
	transitionCanceled  = "cancelled"

	transitionRequested = "requested"
	transitionApproved  = "approved"
	transitionRejected  = "rejected"
	transitionUnhealthy = "unhealthy"

	// transitionAwaitingRetry is emitted on workflow_step.awaiting_retry.v1
	// events when a step has failed and parked waiting for a manual retry,
	// skip, or cancel. Non-terminal: the step's lifecycle event fires later
	// once the workflow unblocks.
	transitionAwaitingRetry = "awaiting_retry"

	// Health transitions. The precise verdict (degraded / unhealthy /
	// healthy) travels in data.metadata.health; the transition says which
	// direction the debounced crossing went. The bad direction reuses
	// transitionUnhealthy above, shared with the runner-unhealthy event.
	transitionRecovered = "recovered"
)

// signalTypeExecuteWorkflow matches the SignalType produced by
// services/ctl-api/internal/pkg/flow/signals/executeflow. Duplicated as a string
// constant to avoid importing the flow package and producing an import cycle.
//
// signalTypeWorkflowStepApprovalRequest / signalTypeWorkflowStepApprovalResponse
// mirror the SignalTypes defined in
// services/ctl-api/internal/app/installs/signals/workflowstepapproval{request,response}.
// Duplicated as string constants for the same reason — to avoid pulling the
// installs/signals tree into the queue/signal/hooks package.
const (
	signalTypeExecuteWorkflow              signal.SignalType = "execute-workflow"
	signalTypeExecuteWorkflowStep          signal.SignalType = "execute-workflow-step"
	signalTypeWorkflowStepApprovalRequest  signal.SignalType = "workflow-step-approval-request"
	signalTypeWorkflowStepApprovalResponse signal.SignalType = "workflow-step-approval-response"
	// signalTypeDriftDetected mirrors driftdetected.SignalType — the
	// notification-only signal dispatched from the plan-only check inside a
	// drift_run / drift_run_reprovision_sandbox workflow when the plan
	// observed actual changes. Its lifecycle events are how subscribers who
	// opted into per-resource `drift_detected: true` get notified.
	signalTypeDriftDetected signal.SignalType = "drift-detected"
	// signalTypeWorkflowStepAwaitingRetry mirrors workflowstepawaitingretry.
	// SignalType — the notification-only signal enqueued when a workflow
	// step fails and parks awaiting manual retry. Its successful after-phase
	// is projected as a workflow_step.awaiting_retry.v1 event with a failed
	// outcome sourced from the step's own error.
	signalTypeWorkflowStepAwaitingRetry signal.SignalType = "workflow-step-awaiting-retry"

	signalTypeStackRun        signal.SignalType = "stack-run"
	signalTypeRoleChange      signal.SignalType = "role-change"
	signalTypeInputsUpdated   signal.SignalType = "inputs-updated"
	signalTypeAppConfigSynced signal.SignalType = "app-config-synced"
	signalTypeUpdateAppConfig signal.SignalType = "update-app-config"
	signalTypeRunnerUnhealthy signal.SignalType = "runner-unhealthy"

	// Component health carriers, mirroring the componenthealthnotify
	// SignalTypes. Emitted by the component-health evaluator on a debounced
	// verdict crossing, outside any workflow — so unlike every signal above
	// they carry no WorkflowID and must be handled before buildEventData's
	// workflow-id guard.
	signalTypeComponentUnhealthy signal.SignalType = "component-unhealthy"
	signalTypeComponentRecovered signal.SignalType = "component-recovered"
	signalTypeInstallDegraded    signal.SignalType = "install-degraded"

	signalTypeSyncInstalls      signal.SignalType = "sync-installs"
	signalTypeInstallConfigSync signal.SignalType = "install-config-sync"
)

// approvalPlanExcerptMaxBytes caps the size of the plan excerpt embedded in
// approval webhook payloads. Slack message limits and consumer log budgets
// make truncation safer than shipping multi-MB plans inline.
const approvalPlanExcerptMaxBytes = 8 * 1024

// orgNameCacheTTL bounds how long a renamed org keeps showing its old name.
const orgNameCacheTTL = 10 * time.Minute

type Params struct {
	fx.In

	Cfg *internal.Config `optional:"true"`
	L   *zap.Logger      `optional:"true"`
	DB  *gorm.DB         `name:"psql" optional:"true"`
	MW  metrics.Writer   `optional:"true"`
}

type WebhookSignalLifecycleHook struct {
	l               *zap.Logger
	httpClient      *http.Client
	webhookURLs     []string
	db              *gorm.DB
	appURL          string
	publicAPIURL    string
	mw              metrics.Writer
	blobReadEnabled bool

	// workflowCreatorCache holds workflowCreatorRow values keyed by workflow
	// id. The underlying row is write-once, so entries never expire.
	workflowCreatorCache sync.Map

	// orgNameCache holds orgNameCacheEntry values keyed by org id. Entries
	// expire after orgNameCacheTTL.
	orgNameCache sync.Map
}

type orgNameCacheEntry struct {
	name      string
	expiresAt time.Time
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
	appURL := ""
	publicAPIURL := ""
	blobReadEnabled := false
	if params.Cfg != nil {
		webhookURLs = params.Cfg.WebhookURLs
		appURL = strings.TrimSpace(params.Cfg.AppURL)
		publicAPIURL = strings.TrimSpace(params.Cfg.PublicAPIURL)
		blobReadEnabled = params.Cfg.BlobReadEnabled
	}

	return &WebhookSignalLifecycleHook{
		l: logger,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		webhookURLs:     normalizeWebhookURLs(webhookURLs),
		db:              params.DB,
		appURL:          appURL,
		publicAPIURL:    publicAPIURL,
		mw:              params.MW,
		blobReadEnabled: blobReadEnabled,
	}
}

// metricNamespace returns the Temporal namespace tag value for metrics emitted
// from inside an activity. Returns "" when called outside an activity context.
func (h *WebhookSignalLifecycleHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

// emitPublishLatency records how long a successful webhook delivery took for
// this phase. Only called when at least one webhook actually fired; lookup-
// only paths and "no targets" exits do not emit so the percentile reflects
// real delivery cost.
func (h *WebhookSignalLifecycleHook) emitPublishLatency(ctx context.Context, phasePrefix string, startTS time.Time) {
	if h.mw == nil {
		return
	}
	h.mw.Timing(
		fmt.Sprintf("signal_lifecycle.%s.webhook.publish_latency", phasePrefix),
		time.Since(startTS),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

// emitError increments the webhook error counter for this phase. One increment
// per failed delivery so the count reflects per-attempt failures.
func (h *WebhookSignalLifecycleHook) emitError(ctx context.Context, phasePrefix string) {
	if h.mw == nil {
		return
	}
	h.mw.Incr(
		fmt.Sprintf("signal_lifecycle.%s.webhook.errors", phasePrefix),
		metrics.ToTags(map[string]string{"namespace": h.metricNamespace(ctx)}),
	)
}

func (h *WebhookSignalLifecycleHook) Name() string {
	return "workflow_lifecycle_webhook"
}

// Supports limits this hook to the public lifecycle primitives:
// execute-workflow (workflow lifecycle), execute-workflow-step (step
// lifecycle), and the approval handshake signals (request / response) which
// are projected as workflow_step.approval.v1 events. Inner-signal events
// (plan/apply, component-deploy, etc.) are deliberately ignored — consumers
// should reason in terms of workflow + step + approval.
func (h *WebhookSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if len(h.webhookURLs) == 0 && h.db == nil {
		return false
	}

	switch event.SignalType {
	case signalTypeExecuteWorkflow,
		signalTypeExecuteWorkflowStep,
		signalTypeWorkflowStepApprovalRequest,
		signalTypeWorkflowStepApprovalResponse,
		signalTypeDriftDetected,
		signalTypeWorkflowStepAwaitingRetry,
		signalTypeStackRun,
		signalTypeRoleChange,
		signalTypeInputsUpdated,
		signalTypeAppConfigSynced,
		signalTypeUpdateAppConfig,
		signalTypeComponentUnhealthy,
		signalTypeComponentRecovered,
		signalTypeInstallDegraded,
		signalTypeRunnerUnhealthy,
		signalTypeSyncInstalls,
		signalTypeInstallConfigSync:
		return true
	default:
		return false
	}
}

func (h *WebhookSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	// Only emit *.started events for the execute phase.
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}

	// Approval signals don't have a "started" semantic: a request has either
	// happened (requested) or it hasn't, and a response is intrinsically
	// terminal (approved / rejected). Drift-detected is a single-shot
	// notification (its Execute is a no-op) — a "started" emission would
	// just produce a duplicate event before the real one. Skip both.
	if suppressesStartedEvent(event.SignalType) {
		return signal.AllowPhaseDecision(), nil
	}

	if err := h.publish(ctx, event, nil); err != nil {
		h.l.Debug("failed to publish workflow lifecycle started webhook", zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

// isApprovalSignalType returns true for the approval handshake signals
// (request / response) which feed the workflow_step.approval.v1 cloud event.
func isApprovalSignalType(t signal.SignalType) bool {
	return t == signalTypeWorkflowStepApprovalRequest ||
		t == signalTypeWorkflowStepApprovalResponse
}

func isNotificationOnlySignalType(t signal.SignalType) bool {
	switch t {
	case signalTypeDriftDetected, signalTypeStackRun, signalTypeRoleChange, signalTypeInputsUpdated, signalTypeAppConfigSynced, signalTypeUpdateAppConfig,
		signalTypeRunnerUnhealthy,
		signalTypeComponentUnhealthy, signalTypeComponentRecovered, signalTypeInstallDegraded,
		signalTypeSyncInstalls, signalTypeInstallConfigSync:
		return true
	}
	return false
}

// isComponentHealthSignalType reports whether the signal is one of the
// component-health notification carriers.
func isComponentHealthSignalType(t signal.SignalType) bool {
	return t == signalTypeComponentUnhealthy ||
		t == signalTypeComponentRecovered ||
		t == signalTypeInstallDegraded
}

// suppressesStartedEvent reports whether a signal type's synthetic "started"
// (before-phase) emission should be skipped. Awaiting-retry is listed here
// but deliberately NOT in isNotificationOnlySignalType: that predicate also
// routes Slack messages to standalone (flat) posts, while awaiting-retry
// should thread into the workflow's existing Slack thread like a step event.
func suppressesStartedEvent(t signal.SignalType) bool {
	return isApprovalSignalType(t) ||
		isNotificationOnlySignalType(t) ||
		t == signalTypeWorkflowStepAwaitingRetry
}

func (h *WebhookSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	// Validation phases never produce a public event for these primitives —
	// validation failures of the workflow / step wrappers surface as a failed
	// execute outcome immediately afterward.
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}

	suppress, err := resolveFlowCompletionOutcome(ctx, h.db, event, &outcome)
	if err != nil {
		return fmt.Errorf("unable to resolve workflow outcome before lifecycle completion: %w", err)
	}
	if suppress {
		return nil
	}

	h.l.Debug("workflow lifecycle webhook after-phase",
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("phase", string(event.Phase)),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("status", string(outcome.Status)),
	)

	return h.publish(ctx, event, &outcome)
}

type workflowStatusRow struct {
	ID        string
	DeletedAt soft_delete.DeletedAt
	Status    app.CompositeStatus `gorm:"type:jsonb;serializer:json"`
}

func (workflowStatusRow) TableName() string {
	return (&app.Workflow{}).TableName()
}

// resolveFlowCompletionOutcome reconciles an execute-workflow completion
// event against the workflow row's domain outcome. Resident flows complete
// their queue signal independently of the workflow row, so the transport
// status can read "success" while the workflow is actually parked
// (failed-pending-retry), errored, or cancelled.
//
// Returns suppress=true when the workflow is parked awaiting retry — the
// re-warmed run emits the real completion later. For terminal error /
// cancelled rows it rewrites outcome in place so the published transition,
// status, and interests classification reflect the domain outcome. On DB
// lookup failure it returns an error and callers must not publish: a
// dropped notification is recoverable noise, a false "succeeded" is not.
func resolveFlowCompletionOutcome(ctx context.Context, db *gorm.DB, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (bool, error) {
	if db == nil ||
		event.SignalType != signalTypeExecuteWorkflow ||
		event.Phase != signal.SignalPhaseExecute ||
		event.WorkflowID == "" ||
		outcome.Status != signal.SignalStatusSuccess {
		return false, nil
	}

	var flw workflowStatusRow
	if err := retryDBRead(ctx, func() error {
		return db.WithContext(ctx).
			Select("status").
			Where(workflowStatusRow{ID: event.WorkflowID}).
			First(&flw).Error
	}); err != nil {
		return false, fmt.Errorf("unable to load workflow status for lifecycle completion: %w", err)
	}

	switch flw.Status.Status {
	case app.StatusFailedPendingRetry:
		return true, nil
	case app.StatusError:
		outcome.Status = signal.SignalStatusError
		outcome.ErrMessage = flw.Status.StatusHumanDescription
		if outcome.ErrMessage == "" {
			outcome.ErrMessage = "workflow failed"
		}
	case app.StatusCancelled:
		outcome.Status = signal.SignalStatusCancelled
		outcome.ErrMessage = flw.Status.StatusHumanDescription
		if outcome.ErrMessage == "" {
			outcome.ErrMessage = "workflow cancelled"
		}
	}

	return false, nil
}

// CloudEvents v1.0 envelope.
type cloudEvent struct {
	SpecVersion     string `json:"specversion"`
	ID              string `json:"id"`
	Type            string `json:"type"`
	Source          string `json:"source"`
	Time            string `json:"time"`
	Subject         string `json:"subject"`
	DataContentType string `json:"datacontenttype"`

	NuonOrgID      string `json:"nuonorgid,omitempty"`
	NuonKind       string `json:"nuonkind,omitempty"`
	NuonTransition string `json:"nuontransition,omitempty"`

	// Interests carries the slug list produced by interests.Classify for
	// this event (resource:..., op:..., outcome:..., event:...). Consumers
	// can route by prefix without re-implementing the classifier. Empty
	// slice is omitted via omitempty.
	Interests []string `json:"interests,omitempty"`

	Data lifecycleEventData `json:"data"`
}

// lifecycleEventData is the public webhook payload. It exposes the workflow
// (always) and optionally the step + approval blocks alongside the transition,
// outcome, and dashboard links.
type lifecycleEventData struct {
	Kind       string `json:"kind"`
	Transition string `json:"transition"`
	OrgID      string `json:"org_id,omitempty"`
	OrgName    string `json:"org_name,omitempty"`

	Workflow workflowRef       `json:"workflow"`
	Step     *workflowStepRef  `json:"step,omitempty"`
	Parent   *parentRef        `json:"parent,omitempty"`
	Outcome  *lifecycleOutcome `json:"outcome,omitempty"`
	Approval *approvalRef      `json:"approval,omitempty"`
	Links    *contextLinks     `json:"links,omitempty"`
	Metadata map[string]any    `json:"metadata,omitempty"`
}

type workflowRef struct {
	ID        string `json:"id"`
	Type      string `json:"type,omitempty"`
	OwnerID   string `json:"owner_id,omitempty"`
	OwnerType string `json:"owner_type,omitempty"`
	// OwnerName is the human-readable name of the workflow owner (e.g. the
	// install name when OwnerType == "installs"). Populated opportunistically
	// when the data is available from existing enrichment JOINs without an
	// extra round-trip; may be empty in other cases.
	OwnerName string `json:"owner_name,omitempty"`
	// CreatedByEmail labels who started the workflow. Falls back to the
	// raw account id for accounts without an email; empty when the
	// workflow has no creator.
	CreatedByEmail string `json:"created_by_email,omitempty"`
	// CreatedAt is the workflow's start time. Zero when unknown.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// RunbookName labels the runbook this workflow is executing. Populated
	// only for runbook_run workflows; sourced from install_workflows.metadata.
	RunbookName string `json:"runbook_name,omitempty"`
}

type workflowStepRef struct {
	ID            string `json:"id"`
	Name          string `json:"name,omitempty"`
	Idx           int    `json:"idx"`
	TargetType    string `json:"target_type,omitempty"`
	TargetID      string `json:"target_id,omitempty"`
	ComponentID   string `json:"component_id,omitempty"`
	ComponentName string `json:"component_name,omitempty"`
	SandboxID     string `json:"sandbox_id,omitempty"`
	ExecutionType string `json:"execution_type,omitempty"`
}

type parentRef struct {
	WorkflowID string `json:"workflow_id,omitempty"`
	StepID     string `json:"step_id,omitempty"`
	Kind       string `json:"kind,omitempty"`
	// ActionName is the human-readable name of the action workflow when the
	// parent step targets an install_action_workflow_run. Empty otherwise.
	ActionName string `json:"action_name,omitempty"`
}

type lifecycleOutcome struct {
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
}

// approvalRef is the public projection of an install workflow step approval.
// The plan is truncated to approvalPlanExcerptMaxBytes; consumers that need
// the full plan should follow the Approval link in contextLinks.
type approvalRef struct {
	ID          string `json:"id"`
	Type        string `json:"type,omitempty"`
	Plan        string `json:"plan,omitempty"`
	RespondedBy string `json:"responded_by,omitempty"`
}

type contextLinks struct {
	Org        string `json:"org,omitempty"`
	Install    string `json:"install,omitempty"`
	Workflow   string `json:"workflow,omitempty"`
	Sandbox    string `json:"sandbox,omitempty"`
	Component  string `json:"component,omitempty"`
	Approval   string `json:"approval,omitempty"`
	RespondAPI string `json:"respond_api,omitempty"`
}

// webhookTarget is a single dispatch target. ConfigOnly is true for entries
// derived from the static h.webhookURLs config list — those have no
// per-target Interests or Match storage and are treated as AllEvents=true /
// org-wide (they receive every supported event in the org). ConfigOnly=false
// targets carry the Interests + Match values loaded from the app.Webhook row.
type webhookTarget struct {
	URL        string
	Secret     string
	ConfigOnly bool
	Interests  interests.Interests
	Match      *labels.SubscriptionMatch
}

func (h *WebhookSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	// outcome is nil when invoked from BeforePhase, non-nil from AfterPhase.
	// Used as the metric-name prefix so before/after timings are split into
	// separate timeseries (see signal_lifecycle.{before,after}_phase.webhook.*).
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

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("workflow_id", event.WorkflowID),
		zap.String("signal_type", string(event.SignalType)),
	)

	// Resolve dispatch targets BEFORE the expensive buildEventData
	// enrichment. listOrgWebhookTargets is a single SELECT on the
	// webhooks table filtered by org_id; the enrichment chain
	// (enrichStep + lookupDeployTargetMeta + lookupParent + approval
	// lookups) issues several JOIN queries and is wasted work when no
	// subscribers exist for this org. Most orgs have no webhooks
	// configured, so this short-circuit removes the dominant DB cost
	// from the activity's hot path.
	//
	// Static config webhooks have no per-target Interests storage; treat
	// them as AllEvents=true so the per-target Matches() filter below
	// passes them through unchanged.
	targets := make([]webhookTarget, 0, len(h.webhookURLs))
	for _, webhookURL := range h.webhookURLs {
		targets = append(targets, webhookTarget{
			URL:        webhookURL,
			ConfigOnly: true,
		})
	}

	dynamicTargets, err := h.listOrgWebhookTargets(ctx, event.OrgID)
	if err != nil {
		logger.Warn("failed to resolve org workflow lifecycle webhooks", zap.Error(err))
	}

	targets = append(targets, dynamicTargets...)
	targets = dedupeWebhookTargets(targets)
	if len(targets) == 0 {
		return nil
	}

	data, ok := h.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	ceType := cloudEventTypeWorkflow
	switch data.Kind {
	case kindWorkflowStep:
		ceType = cloudEventTypeWorkflowStep
	case kindWorkflowStepApproval:
		ceType = cloudEventTypeWorkflowStepApproval
	case kindStackRun:
		ceType = cloudEventTypeStackRun
	case kindRoleChange:
		ceType = cloudEventTypeRoleChange
	case kindInputsUpdated:
		ceType = cloudEventTypeInputsUpdated
	case kindAppConfigSynced:
		ceType = cloudEventTypeAppConfigSynced
	case kindUpdateAppConfig:
		ceType = cloudEventTypeUpdateAppConfig
	case kindRunnerUnhealthy:
		ceType = cloudEventTypeRunnerUnhealthy
	case kindComponentHealth:
		ceType = cloudEventTypeComponentHealth
	case kindInstallHealth:
		ceType = cloudEventTypeInstallHealth
	case kindInstallSync:
		ceType = cloudEventTypeInstallSync
	case kindInstallConfigSync:
		ceType = cloudEventTypeInstallConfigSync
	}
	// Awaiting-retry shares kind=workflow_step with the normal step
	// lifecycle but gets its own CloudEvent type so consumers can route the
	// "action required" case without parsing the transition.
	if event.SignalType == signalTypeWorkflowStepAwaitingRetry {
		ceType = cloudEventTypeWorkflowStepAwaitingRetry
	}

	subject := buildSubject(event, data)

	// Slug list mirrors the per-target Matches() decision so consumers can
	// route by prefix (resource:installs, op:components.deploy,
	// outcome:failures, event:lifecycle.failed, etc.) without
	// re-implementing the classifier.
	slugs := interests.Classify(event, outcome, h.db)

	ce := cloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Type:            ceType,
		Source:          "//nuon.co/ctl-api",
		Time:            time.Now().UTC().Format(time.RFC3339),
		Subject:         subject,
		DataContentType: "application/json",
		NuonOrgID:       event.OrgID,
		NuonKind:        data.Kind,
		NuonTransition:  data.Transition,
		Interests:       slugs,
		Data:            data,
	}

	payloadJSON, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("unable to marshal workflow lifecycle webhook payload: %w", err)
	}

	logger = logger.With(
		zap.String("kind", data.Kind),
		zap.String("transition", data.Transition),
		zap.Int("webhook_count", len(targets)),
	)

	// Resolve the entity ids referenced by this event (install / component /
	// action). Drives the per-target Match.Matches predicate below. Any of
	// these may be empty for org-only events; nil-Match targets (org-wide)
	// always fire regardless.
	matchTargets := EventTargetsFromEvent(ctx, h.db, event, data)

	// labelLoader memoises label lookups for this publish() call. Multiple
	// targets in the same org that hit the same install / component / action
	// only pay the SELECT cost once. Local to the call (events fan out
	// across many publish() invocations for unrelated workflows).
	labelLoader := newLabelLoader(h.db)

	var sendErrs []error
	for _, target := range targets {
		// Per-target Match filter. ConfigOnly targets bypass the
		// predicate (always org-wide). nil Match means org-wide and the
		// matcher returns true unconditionally.
		if !target.ConfigOnly && target.Match != nil {
			if err := labelLoader.load(ctx, &matchTargets); err != nil {
				logger.Warn("failed to load event labels for match",
					zap.Error(err))
				// Fail open: a label lookup failure shouldn't drop
				// the dispatch. Selector matches will simply miss
				// when the label set is empty.
			}
			if !target.Match.Matches(matchTargets) {
				continue
			}
		}

		// Per-target interests filter. ConfigOnly targets bypass the
		// matcher (treated as AllEvents=true) since the static config
		// list has no Interests storage.
		if !target.ConfigOnly && !interests.Matches(event, outcome, h.db, target.Interests) {
			continue
		}

		if err := h.sendWebhook(ctx, target, payloadJSON); err != nil {
			sendErrs = append(sendErrs, err)
			h.emitError(ctx, phasePrefix)
			logger.Warn("failed to deliver workflow lifecycle webhook",
				zap.String("webhook_host", webhookHost(target.URL)),
				zap.Error(err))
			continue
		}
		delivered = true
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}

	logger.Debug("delivered workflow lifecycle webhook")
	return nil
}

// buildEventData translates an internal SignalPhaseEvent into the public
// workflow / workflow_step / workflow_step_approval payload. Returns ok=false
// when there is nothing to emit (e.g. missing identifiers).
// buildEventData projects a signal phase event into the public payload shape,
// then backfills the org display name for every builder path. The name is only
// stamped on the event by signals that originate inside a workflow; carrier
// signals (component health, runner unhealthy) have none, and a notification
// that identifies the org by a truncated id reads as an internal log line.
func (h *WebhookSignalLifecycleHook) buildEventData(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	data, ok := h.buildEventDataForSignal(ctx, event, outcome)
	if !ok {
		return data, false
	}
	if data.OrgName == "" {
		data.OrgName = h.lookupOrgName(ctx, event.OrgID)
	}
	return data, true
}

func (h *WebhookSignalLifecycleHook) buildEventDataForSignal(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	switch event.SignalType {
	case signalTypeStackRun, signalTypeRoleChange, signalTypeInputsUpdated:
		return h.buildStackEventData(ctx, event, outcome)
	case signalTypeAppConfigSynced:
		return h.buildAppConfigSyncedEventData(ctx, event, outcome)
	case signalTypeUpdateAppConfig:
		return h.buildUpdateAppConfigEventData(ctx, event, outcome)
	case signalTypeRunnerUnhealthy:
		return h.buildRunnerUnhealthyEventData(event, outcome)
	case signalTypeComponentUnhealthy, signalTypeComponentRecovered, signalTypeInstallDegraded:
		return h.buildComponentHealthEventData(event, outcome)
	case signalTypeSyncInstalls, signalTypeInstallConfigSync:
		return h.buildInstallSyncEventData(event, outcome)
	}

	if event.WorkflowID == "" {
		return lifecycleEventData{}, false
	}

	if isApprovalSignalType(event.SignalType) {
		return h.buildApprovalEventData(ctx, event, outcome)
	}

	// The awaiting-retry carrier only produces a public event from its
	// successful after-phase: the before-phase is suppressed (see
	// suppressesStartedEvent) and a failed/cancelled carrier means the
	// notification itself broke — not something subscribers should see.
	if event.SignalType == signalTypeWorkflowStepAwaitingRetry &&
		(outcome == nil || outcome.Status != signal.SignalStatusSuccess) {
		return lifecycleEventData{}, false
	}

	kind := kindWorkflow
	if event.SignalType == signalTypeExecuteWorkflowStep ||
		event.SignalType == signalTypeWorkflowStepAwaitingRetry {
		kind = kindWorkflowStep
	}

	transition := mapTransition(event, outcome)
	if event.SignalType == signalTypeWorkflowStepAwaitingRetry {
		transition = transitionAwaitingRetry
	}

	data := lifecycleEventData{
		Kind:       kind,
		Transition: transition,
		OrgID:      event.OrgID,
		// OrgName / Workflow.OwnerName are stamped onto the event by the
		// originating signal at Validate() time (see executeflow.Signal.
		// LifecycleContext) and propagated here without a DB lookup.
		OrgName: event.OrgName,
		Workflow: workflowRef{
			ID:        event.WorkflowID,
			Type:      event.WorkflowType,
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
	}

	creator := h.lookupWorkflowCreator(ctx, event.WorkflowID)
	data.Workflow.CreatedByEmail = creator.CreatedByEmail
	data.Workflow.CreatedAt = creator.CreatedAt
	data.Workflow.RunbookName = creator.RunbookName

	if outcome != nil {
		data.Outcome = h.buildOutcome(event, outcome)
	}

	// The awaiting-retry carrier signal itself always succeeds (its Execute
	// is a no-op) — the outcome subscribers care about is the parked step's
	// real failure, carried in the signal's lifecycle metadata. Project it
	// as a failed outcome and pass the metadata (retry_index, max_retries,
	// awaiting_retry / manual_action_required flags) through verbatim.
	if event.SignalType == signalTypeWorkflowStepAwaitingRetry {
		data.Metadata = event.Metadata
		if data.Outcome != nil {
			data.Outcome.Status = statusFailed
			if errMsg, _ := event.Metadata["error"].(string); errMsg != "" {
				data.Outcome.Error = errMsg
			}
		}
	}

	// Enrich the step on workflow_step.lifecycle events AND on the
	// drift-detected signal — the latter has Kind=workflow but its StepID
	// points at the plan-only step that observed the drift, and the renderer
	// (Slack flat drift message + webhook subscribers) needs ComponentName /
	// SandboxID and the matching Component / Sandbox dashboard links to
	// render a useful "drift detected on X" payload.
	if event.StepID != "" && (kind == kindWorkflowStep || event.SignalType == signalTypeDriftDetected) {
		stepRef, installName, emit := h.enrichStep(ctx, event.StepID)
		// Drop the entire step.lifecycle event for hidden / internal steps
		// (e.g. "generate install state"). Webhook consumers should only see
		// user-facing steps; system bookkeeping steps are filtered here.
		if !emit {
			return lifecycleEventData{}, false
		}
		data.Step = stepRef
		// Fallback: if the step JOIN surfaced an install name and the event
		// itself didn't carry one (older signals enqueued before owner_name
		// stamping shipped), use it. New workflow runs will already have
		// data.Workflow.OwnerName populated from the event.
		if installName != "" && data.Workflow.OwnerType == "installs" && data.Workflow.OwnerName == "" {
			data.Workflow.OwnerName = installName
		}
	}

	data.Parent = h.lookupParent(ctx, event.WorkflowID)

	data.Links = h.buildContextLinks(event, data.Step)

	return data, true
}

func (h *WebhookSignalLifecycleHook) buildStackEventData(_ context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	var kind string
	switch event.SignalType {
	case signalTypeStackRun:
		kind = kindStackRun
	case signalTypeRoleChange:
		kind = kindRoleChange
	case signalTypeInputsUpdated:
		kind = kindInputsUpdated
	default:
		return lifecycleEventData{}, false
	}

	transition := mapTransition(event, outcome)

	data := lifecycleEventData{
		Kind:       kind,
		Transition: transition,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Metadata: event.Metadata,
	}

	if outcome != nil {
		data.Outcome = h.buildOutcome(event, outcome)
	}

	return data, true
}

func (h *WebhookSignalLifecycleHook) buildAppConfigSyncedEventData(_ context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	transition := mapTransition(event, outcome)
	data := lifecycleEventData{
		Kind:       kindAppConfigSynced,
		Transition: transition,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Metadata: event.Metadata,
	}
	if outcome != nil {
		data.Outcome = h.buildOutcome(event, outcome)
	}
	return data, true
}

func (h *WebhookSignalLifecycleHook) buildInstallSyncEventData(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	kind := kindInstallSync
	if event.SignalType == signalTypeInstallConfigSync {
		kind = kindInstallConfigSync
	}
	transition := mapTransition(event, outcome)
	data := lifecycleEventData{
		Kind:       kind,
		Transition: transition,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Metadata: event.Metadata,
	}
	if outcome != nil {
		data.Outcome = h.buildOutcome(event, outcome)
	}
	return data, true
}

func (h *WebhookSignalLifecycleHook) buildUpdateAppConfigEventData(_ context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	transition := mapTransition(event, outcome)
	data := lifecycleEventData{
		Kind:       kindUpdateAppConfig,
		Transition: transition,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Metadata: event.Metadata,
	}
	if outcome != nil {
		data.Outcome = h.buildOutcome(event, outcome)
	}
	return data, true
}

// buildComponentHealthEventData projects a component-health carrier into the
// public payload. There is no workflow or step behind these events — the
// component and install identity plus the render fields (health,
// previous_health, message, failing resource) arrive on the signal's lifecycle
// context and pass straight through as metadata.
//
// Only the carrier's successful after-phase produces an event: a failed
// carrier means the notification plumbing broke, which is not a health
// transition subscribers should see.
func (h *WebhookSignalLifecycleHook) buildComponentHealthEventData(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	if outcome == nil || outcome.Status != signal.SignalStatusSuccess {
		return lifecycleEventData{}, false
	}

	kind := kindComponentHealth
	if event.SignalType == signalTypeInstallDegraded {
		kind = kindInstallHealth
	}

	transition := transitionUnhealthy
	if event.SignalType == signalTypeComponentRecovered {
		transition = transitionRecovered
	}
	if event.SignalType == signalTypeInstallDegraded && !isBadHealthMetadata(event.Metadata) {
		transition = transitionRecovered
	}

	data := lifecycleEventData{
		Kind:       kind,
		Transition: transition,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Metadata: event.Metadata,
	}

	// buildContextLinks derives the component deep link from a step's
	// ComponentID. These events have no step, so pass a link-only stand-in —
	// data.Step stays nil so the payload doesn't claim a step that never ran.
	var linkStep *workflowStepRef
	if event.ComponentID != nil && *event.ComponentID != "" {
		linkStep = &workflowStepRef{ComponentID: *event.ComponentID}
	}
	data.Links = h.buildContextLinks(event, linkStep)

	return data, true
}

// isBadHealthMetadata reports whether the carrier's health metadata describes
// a problem state, used to label the install-level transition direction.
func isBadHealthMetadata(metadata map[string]any) bool {
	health, _ := metadata["health"].(string)
	return health == "degraded" || health == "unhealthy"
}

func (h *WebhookSignalLifecycleHook) buildRunnerUnhealthyEventData(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	if event.Phase != signal.SignalPhaseExecute || outcome == nil || outcome.Status != signal.SignalStatusSuccess {
		return lifecycleEventData{}, false
	}

	reason, _ := event.Metadata["reason"].(string)
	data := lifecycleEventData{
		Kind:       kindRunnerUnhealthy,
		Transition: transitionUnhealthy,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Outcome: &lifecycleOutcome{
			Status: statusFailed,
			Error:  reason,
		},
		Metadata: event.Metadata,
	}
	data.Links = h.buildContextLinks(event, nil)
	return data, true
}

func (h *WebhookSignalLifecycleHook) buildOutcome(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) *lifecycleOutcome {
	out := &lifecycleOutcome{
		Status: mapStatus(outcome.Status),
	}
	if outcome.Status != signal.SignalStatusSuccess {
		out.Error = outcome.ErrMessage
	}
	if outcome.Duration > 0 {
		out.DurationMs = outcome.Duration.Milliseconds()
	}
	if event.Phase == signal.SignalPhaseCancel {
		out.Status = statusCanceled
	}
	return out
}

// buildApprovalEventData projects an approval-request / approval-response
// signal event into a workflow_step.approval.v1 payload. We only emit on a
// successful execute outcome — failures of the wrapper signal itself aren't
// part of the public approval vocabulary, and consumers care about the
// approval state transition, not the bookkeeping signal's lifecycle.
func (h *WebhookSignalLifecycleHook) buildApprovalEventData(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	if event.StepID == "" || h.db == nil {
		return lifecycleEventData{}, false
	}
	// Approval events are only emitted after the underlying signal's execute
	// phase has succeeded. We deliberately drop validation, before-phase, and
	// failure events for these signals.
	if event.Phase != signal.SignalPhaseExecute || outcome == nil ||
		outcome.Status != signal.SignalStatusSuccess {
		return lifecycleEventData{}, false
	}

	approval, ok := h.lookupStepApproval(ctx, event.StepID)
	if !ok {
		return lifecycleEventData{}, false
	}

	var (
		transition  string
		respondedBy string
	)
	switch event.SignalType {
	case signalTypeWorkflowStepApprovalRequest:
		transition = transitionRequested
	case signalTypeWorkflowStepApprovalResponse:
		responseRow, found := h.lookupApprovalResponse(ctx, approval.ID)
		if !found {
			return lifecycleEventData{}, false
		}
		transition = mapApprovalResponseTransition(responseRow.Type)
		if transition == "" {
			// Retry / unknown response types don't surface as approval
			// transitions — retry is handled by the step-group, not the
			// approval vocabulary.
			return lifecycleEventData{}, false
		}
		respondedBy = responseRow.RespondedBy
	default:
		return lifecycleEventData{}, false
	}

	stepRef, installName, emit := h.enrichStep(ctx, event.StepID)
	if !emit {
		return lifecycleEventData{}, false
	}

	data := lifecycleEventData{
		Kind:       kindWorkflowStepApproval,
		Transition: transition,
		OrgID:      event.OrgID,
		OrgName:    event.OrgName,
		Workflow: workflowRef{
			ID:        event.WorkflowID,
			Type:      event.WorkflowType,
			OwnerID:   event.OwnerID,
			OwnerType: event.OwnerType,
			OwnerName: event.OwnerName,
		},
		Step: stepRef,
		Approval: &approvalRef{
			ID:          approval.ID,
			Type:        string(approval.Type),
			Plan:        truncateApprovalPlan(h.approvalContents(ctx, approval)),
			RespondedBy: respondedBy,
		},
	}

	creator := h.lookupWorkflowCreator(ctx, event.WorkflowID)
	data.Workflow.CreatedByEmail = creator.CreatedByEmail
	data.Workflow.CreatedAt = creator.CreatedAt
	data.Workflow.RunbookName = creator.RunbookName

	if data.OrgName == "" {
		data.OrgName = h.lookupOrgName(ctx, event.OrgID)
	}

	if installName != "" && data.Workflow.OwnerType == "installs" && data.Workflow.OwnerName == "" {
		data.Workflow.OwnerName = installName
	}

	data.Parent = h.lookupParent(ctx, event.WorkflowID)
	data.Links = h.buildContextLinks(event, data.Step)
	if data.Links != nil {
		// The dashboard SPA does not currently serve a per-step or
		// per-approval route — step detail renders inline on the workflow
		// page (see services/dashboard-ui/client/views/install/routes.tsx).
		// Point links.approval at the workflow page so consumers land on a
		// real, working URL where the approval is visible. When the
		// dashboard adds a real /steps/:stepId/approvals/:approvalId route
		// this can grow back into a true deep link without a wire-format
		// change.
		data.Links.Approval = data.Links.Workflow
		data.Links.RespondAPI = h.respondAPIURL(event.WorkflowID, event.StepID, approval.ID)
	}

	return data, true
}

// mapApprovalResponseTransition translates a WorkflowStepResponseType into
// the public approval transition vocabulary. Retry is intentionally excluded
// — the create-approval-response handler routes retry through RetryStep
// (see services/ctl-api/internal/app/installs/service/create_workflow_step_approval_response.go),
// so the approval-response signal never carries a retry response type.
func mapApprovalResponseTransition(t app.WorkflowStepResponseType) string {
	switch t {
	case app.WorkflowStepApprovalResponseTypeApprove,
		app.WorkflowStepApprovalResponseTypeAutoApprove:
		return transitionApproved
	case app.WorkflowStepApprovalResponseTypeDeny,
		app.WorkflowStepApprovalResponseTypeSkipCurrent,
		app.WorkflowStepApprovalResponseTypeSkipCurrentAndDependents:
		return transitionRejected
	default:
		return ""
	}
}

// truncateApprovalPlan trims a plan blob to approvalPlanExcerptMaxBytes,
// appending an explicit "(truncated)" marker when truncation occurred so the
// receiving consumer can render the right indicator.
// approvalContents reads the approval plan, honoring the blob-read flag, and
// logs when the read is served from the S3 blob.
func (h *WebhookSignalLifecycleHook) approvalContents(ctx context.Context, approval *app.WorkflowStepApproval) string {
	contents, fromBlob := approval.GetContents(ctx, h.blobReadEnabled)
	if fromBlob {
		h.l.Debug("read workflow step approval contents from blob",
			zap.String("approval_id", approval.ID),
			zap.Int("bytes", len(contents)))
	}
	return contents
}

func truncateApprovalPlan(plan string) string {
	plan = strings.TrimSpace(plan)
	if plan == "" {
		return ""
	}
	if len(plan) <= approvalPlanExcerptMaxBytes {
		return plan
	}
	return plan[:approvalPlanExcerptMaxBytes] + "\n... (truncated)"
}

// lookupStepApproval finds the (single) un-deleted approval row attached to a
// workflow step. Returns ok=false when no approval exists or the lookup
// fails.
func (h *WebhookSignalLifecycleHook) lookupStepApproval(ctx context.Context, stepID string) (*app.WorkflowStepApproval, bool) {
	if h.db == nil || stepID == "" {
		return nil, false
	}
	var approval app.WorkflowStepApproval
	if err := h.db.WithContext(ctx).
		Where("install_workflow_step_id = ?", stepID).
		Order("created_at DESC").
		First(&approval).Error; err != nil {
		h.l.Debug("failed to load workflow step approval for webhook enrichment",
			zap.String("step_id", stepID),
			zap.Error(err))
		return nil, false
	}
	return &approval, true
}

// approvalResponseRow carries the response type plus the responder's display
// label resolved by joining accounts to the response row.
type approvalResponseRow struct {
	Type        app.WorkflowStepResponseType
	RespondedBy string
}

// lookupApprovalResponse fetches the response attached to an approval and the
// human-readable identity of the responder. Best-effort: returns found=false
// when the response can't be located.
func (h *WebhookSignalLifecycleHook) lookupApprovalResponse(ctx context.Context, approvalID string) (approvalResponseRow, bool) {
	if h.db == nil || approvalID == "" {
		return approvalResponseRow{}, false
	}
	var row struct {
		ResponseType string
		RespondedBy  string
	}
	if err := h.db.WithContext(ctx).
		Table("install_workflow_step_approval_responses AS r").
		Select(`r.type AS response_type,
			COALESCE(NULLIF(acc.email, ''), acc.id, '') AS responded_by`).
		Joins("LEFT JOIN accounts AS acc ON acc.id = r.created_by_id").
		Where("r.install_workflow_step_approval_id = ?", approvalID).
		Order("r.created_at DESC").
		Limit(1).
		Scan(&row).Error; err != nil {
		h.l.Debug("failed to load workflow step approval response for webhook enrichment",
			zap.String("approval_id", approvalID),
			zap.Error(err))
		return approvalResponseRow{}, false
	}
	if row.ResponseType == "" {
		return approvalResponseRow{}, false
	}
	return approvalResponseRow{
		Type:        app.WorkflowStepResponseType(row.ResponseType),
		RespondedBy: row.RespondedBy,
	}, true
}

// lookupOrgName returns the org's display name. Empty when the row is
// missing or the lookup fails.
func (h *WebhookSignalLifecycleHook) lookupOrgName(ctx context.Context, orgID string) string {
	if h.db == nil || orgID == "" {
		return ""
	}
	if v, ok := h.orgNameCache.Load(orgID); ok {
		entry := v.(orgNameCacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.name
		}
	}
	var name string
	if err := h.db.WithContext(ctx).
		Table("orgs").
		Select("name").
		Where("id = ?", orgID).
		Limit(1).
		Scan(&name).Error; err != nil {
		h.l.Debug("failed to load org name for lifecycle enrichment",
			zap.String("org_id", orgID),
			zap.Error(err))
		return ""
	}
	h.orgNameCache.Store(orgID, orgNameCacheEntry{
		name:      name,
		expiresAt: time.Now().Add(orgNameCacheTTL),
	})
	return name
}

type workflowCreatorRow struct {
	CreatedByEmail string
	CreatedAt      time.Time
	RunbookName    string
}

// lookupWorkflowCreator returns who started the workflow and when. Email
// when the account has one, raw account id otherwise. Zero value when the
// row is missing or the lookup fails.
func (h *WebhookSignalLifecycleHook) lookupWorkflowCreator(ctx context.Context, workflowID string) workflowCreatorRow {
	if h.db == nil || workflowID == "" {
		return workflowCreatorRow{}
	}
	if v, ok := h.workflowCreatorCache.Load(workflowID); ok {
		return v.(workflowCreatorRow)
	}
	var row workflowCreatorRow
	if err := h.db.WithContext(ctx).
		Table("install_workflows AS w").
		Select(`COALESCE(NULLIF(acc.email, ''), w.created_by_id, '') AS created_by_email,
			w.created_at AS created_at,
			COALESCE(w.metadata->'runbook_name', '') AS runbook_name`).
		Joins("LEFT JOIN accounts AS acc ON acc.id = w.created_by_id").
		Where("w.id = ?", workflowID).
		Limit(1).
		Scan(&row).Error; err != nil {
		h.l.Debug("failed to load workflow creator for lifecycle enrichment",
			zap.String("workflow_id", workflowID),
			zap.Error(err))
		return workflowCreatorRow{}
	}
	h.workflowCreatorCache.Store(workflowID, row)
	return row
}

// respondAPIURL builds the ctl-api endpoint a consumer would POST to in order
// to create an approval response. Returns "" when PublicAPIURL is
// unconfigured.
//
// NOTE: today this is wire-format only. Slackbot doesn't yet have an outbound
// HTTP client to ctl-api or an authenticated identity to act on a user's
// behalf, so the URL is provided so future Approve/Reject buttons can wire
// directly without a payload-shape change.
func (h *WebhookSignalLifecycleHook) respondAPIURL(workflowID, stepID, approvalID string) string {
	if h.publicAPIURL == "" || workflowID == "" || stepID == "" || approvalID == "" {
		return ""
	}
	link, err := url.JoinPath(h.publicAPIURL,
		"v1", "workflows", workflowID,
		"steps", stepID,
		"approvals", approvalID,
		"response",
	)
	if err != nil {
		return ""
	}
	return link
}

// enrichStep loads the workflow step by id and projects the user-facing step
// fields. This is server-side and runs in an activity context, so DB access
// is safe and replay-deterministic.
//
// Returns the step ref plus the resolved install name (when the step's target
// allows it to be discovered cheaply via the same JOIN). The install name is
// returned out-of-band so the caller can place it on workflow.owner_name.
//
// The third return value (emit) is false when the step is internal /
// system-only (ExecutionType == "hidden", e.g. "generate install state") and
// the entire step.lifecycle event should be suppressed. On DB errors we
// fail open (emit=true) so transient infra issues don't silently drop user
// events.
func (h *WebhookSignalLifecycleHook) enrichStep(ctx context.Context, stepID string) (*workflowStepRef, string, bool) {
	ref := &workflowStepRef{ID: stepID}
	if h.db == nil {
		return ref, "", true
	}

	var step app.WorkflowStep
	if err := h.db.WithContext(ctx).
		Where("id = ?", stepID).
		First(&step).Error; err != nil {
		h.l.Debug("failed to load workflow step for webhook enrichment",
			zap.String("step_id", stepID),
			zap.Error(err))
		return ref, "", true
	}

	if step.ExecutionType == app.WorkflowStepExecutionTypeHidden {
		return ref, "", false
	}

	ref.Name = step.Name
	ref.Idx = step.Idx
	ref.TargetType = step.StepTargetType
	ref.TargetID = step.StepTargetID
	ref.ExecutionType = string(step.ExecutionType)

	// Both singular and plural target type strings exist in the codebase
	// (WorkflowStepTargetTypeInstallDeploy / *Deploys, etc.) but the actual
	// install_workflow_steps.step_target_type column is consistently the
	// plural form ("install_deploys", "install_sandbox_runs"). Match both
	// defensively so any legacy singular row also enriches correctly.
	var installName string
	switch step.StepTargetType {
	case string(app.WorkflowStepTargetTypeInstallDeploy),
		string(app.WorkflowStepTargetTypeInstallDeploys):
		meta := h.lookupDeployTargetMeta(ctx, step.StepTargetID)
		ref.ComponentID = meta.ComponentID
		ref.ComponentName = meta.ComponentName
		installName = meta.InstallName
	case string(app.WorkflowStepTargetTypeInstallSandboxRun),
		string(app.WorkflowStepTargetTypeInstallSandboxRuns):
		meta := h.lookupSandboxRunTargetMeta(ctx, step.StepTargetID)
		ref.SandboxID = meta.SandboxID
		installName = meta.InstallName
	}

	return ref, installName, true
}

// deployTargetMeta carries the install/component identity & names resolved for
// an install_deploys step target via a single JOIN.
type deployTargetMeta struct {
	ComponentID   string
	ComponentName string
	InstallName   string
}

// lookupDeployTargetMeta resolves component id/name and install name for an
// install deploy target. Best-effort: returns a zero struct on any DB error.
func (h *WebhookSignalLifecycleHook) lookupDeployTargetMeta(ctx context.Context, deployID string) deployTargetMeta {
	if h.db == nil || deployID == "" {
		return deployTargetMeta{}
	}
	var row struct {
		ComponentID   string
		ComponentName string
		InstallName   string
	}
	if err := h.db.WithContext(ctx).
		Table("install_deploys").
		Select(`install_components.component_id AS component_id,
			components.name AS component_name,
			installs.name AS install_name`).
		Joins("JOIN install_components ON install_components.id = install_deploys.install_component_id").
		Joins("LEFT JOIN components ON components.id = install_components.component_id").
		Joins("LEFT JOIN installs ON installs.id = install_components.install_id").
		Where("install_deploys.id = ?", deployID).
		Scan(&row).Error; err != nil {
		return deployTargetMeta{}
	}
	return deployTargetMeta{
		ComponentID:   row.ComponentID,
		ComponentName: row.ComponentName,
		InstallName:   row.InstallName,
	}
}

// sandboxRunTargetMeta carries the sandbox id & install name resolved for an
// install_sandbox_runs step target.
type sandboxRunTargetMeta struct {
	SandboxID   string
	InstallName string
}

// lookupSandboxRunTargetMeta resolves sandbox id and install name for an
// install sandbox run target.
func (h *WebhookSignalLifecycleHook) lookupSandboxRunTargetMeta(ctx context.Context, sandboxRunID string) sandboxRunTargetMeta {
	if h.db == nil || sandboxRunID == "" {
		return sandboxRunTargetMeta{}
	}
	var row struct {
		SandboxID   string
		InstallName string
	}
	if err := h.db.WithContext(ctx).
		Table("install_sandbox_runs").
		Select(`install_sandbox_runs.install_sandbox_id AS sandbox_id,
			installs.name AS install_name`).
		Joins("LEFT JOIN installs ON installs.id = install_sandbox_runs.install_id").
		Where("install_sandbox_runs.id = ?", sandboxRunID).
		Scan(&row).Error; err != nil {
		return sandboxRunTargetMeta{}
	}
	return sandboxRunTargetMeta{
		SandboxID:   row.SandboxID,
		InstallName: row.InstallName,
	}
}

// lookupParent resolves a parent {workflow_id, step_id, kind} block when this
// workflow is nested inside another workflow's step (e.g. an action workflow
// run launched from a deploy step). Returns nil when no parent is detected.
func (h *WebhookSignalLifecycleHook) lookupParent(ctx context.Context, workflowID string) *parentRef {
	if h.db == nil || workflowID == "" {
		return nil
	}

	// Action workflow runs link back to their parent workflow via
	// install_action_workflow_runs.install_workflow_id. The parent step is the
	// step in that parent workflow whose target points at the run.
	//
	// We also LEFT JOIN through install_action_workflows → action_workflows to
	// pick up the human-readable action name. The joins are LEFT because the
	// run's install_action_workflow_id is nullable (manual triggers may not
	// reference a stored action workflow).
	var row struct {
		ParentWorkflowID string
		ParentStepID     string
		ActionName       string
	}
	err := h.db.WithContext(ctx).
		Raw(`
			SELECT
				iws.install_workflow_id AS parent_workflow_id,
				iws.id AS parent_step_id,
				aw.name AS action_name
			FROM install_action_workflow_runs iawr
			JOIN install_workflow_steps iws
			  ON iws.step_target_type = ?
			 AND iws.step_target_id = iawr.id
			LEFT JOIN install_action_workflows iaw
			  ON iaw.id = iawr.install_action_workflow_id
			LEFT JOIN action_workflows aw
			  ON aw.id = iaw.action_workflow_id
			WHERE iawr.install_workflow_id = ?
			LIMIT 1`,
			string(app.WorkflowStepTargetTypeInstallActionWorkflowRun),
			workflowID,
		).Scan(&row).Error

	if err != nil || row.ParentWorkflowID == "" {
		return nil
	}

	return &parentRef{
		WorkflowID: row.ParentWorkflowID,
		StepID:     row.ParentStepID,
		Kind:       kindWorkflowStep,
		ActionName: row.ActionName,
	}
}

// mapTransition derives the public transition string from the phase + outcome.
// Workflow / step events emit:
//   - started   on BeforePhase(execute)
//   - succeeded on AfterPhase(execute) success
//   - failed    on AfterPhase(execute) error
//   - cancelled on AfterPhase(cancel)
func mapTransition(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) string {
	if outcome == nil {
		return transitionStarted
	}
	if event.Phase == signal.SignalPhaseCancel {
		return transitionCanceled
	}
	switch outcome.Status {
	case signal.SignalStatusSuccess:
		return transitionSucceeded
	case signal.SignalStatusCancelled:
		return transitionCanceled
	default:
		return transitionFailed
	}
}

// mapStatus converts the internal SignalStatus into the user-facing status
// string used in webhook payloads.
func mapStatus(s signal.SignalStatus) string {
	switch s {
	case signal.SignalStatusSuccess:
		return statusSucceeded
	case signal.SignalStatusError:
		return statusFailed
	case signal.SignalStatusCancelled:
		return statusCanceled
	default:
		return string(s)
	}
}

// buildSubject returns a stable identifier used as the CloudEvent `subject`.
// Composed from org id, kind, workflow id, and (when applicable) step id so
// consumers can correlate started/finished pairs without parsing payloads.
func buildSubject(event signal.SignalPhaseEvent, data lifecycleEventData) string {
	parts := []string{}
	if event.OrgID != "" {
		parts = append(parts, event.OrgID)
	}
	parts = append(parts, data.Kind)
	if data.Kind == kindRunnerUnhealthy {
		if runnerID, _ := data.Metadata["runner_id"].(string); runnerID != "" {
			parts = append(parts, runnerID)
		}
	}
	if data.Workflow.ID != "" {
		parts = append(parts, data.Workflow.ID)
	}
	if data.Step != nil && data.Step.ID != "" {
		parts = append(parts, data.Step.ID)
	}
	parts = append(parts, data.Transition)
	return strings.Join(parts, "/")
}

func (h *WebhookSignalLifecycleHook) sendWebhook(ctx context.Context, target webhookTarget, payloadJSON []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target.URL, bytes.NewReader(payloadJSON))
	if err != nil {
		return fmt.Errorf("unable to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/cloudevents+json; charset=utf-8")
	if target.Secret != "" {
		mac := hmac.New(sha256.New, []byte(target.Secret))
		mac.Write(payloadJSON)
		req.Header.Set("X-Nuon-Signature", hex.EncodeToString(mac.Sum(nil)))
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to execute webhook request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4*1024))
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		return fmt.Errorf("unable to drain webhook response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		if len(body) == 0 {
			return fmt.Errorf("webhook endpoint returned status %d", resp.StatusCode)
		}

		return fmt.Errorf("webhook endpoint returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	return nil
}

func (h *WebhookSignalLifecycleHook) listOrgWebhookTargets(ctx context.Context, orgID string) ([]webhookTarget, error) {
	if h.db == nil || orgID == "" {
		return nil, nil
	}

	var webhooks []app.Webhook
	if err := retryDBRead(ctx, func() error {
		return h.db.WithContext(ctx).
			Where("org_id = ?", orgID).
			Find(&webhooks).Error
	}); err != nil {
		return nil, fmt.Errorf("unable to list org workflow lifecycle webhooks: %w", err)
	}

	targets := make([]webhookTarget, 0, len(webhooks))
	for _, webhook := range webhooks {
		trimmedURL := strings.TrimSpace(webhook.WebhookURL)
		if trimmedURL == "" {
			continue
		}

		// Treat an unconfigured (zero-value) Interests as "all events".
		// Webhooks predate the interests filter, so any row whose JSONB
		// column is NULL/empty must keep receiving everything by default.
		// New rows get AllEvents() at create time in the service layer.
		effectiveInterests := webhook.Interests
		if effectiveInterests.IsZero() {
			effectiveInterests = interests.AllEvents()
		}

		targets = append(targets, webhookTarget{
			URL:       trimmedURL,
			Secret:    strings.TrimSpace(webhook.WebhookSecret),
			Interests: effectiveInterests,
			Match:     webhook.Match,
		})
	}

	return targets, nil
}

// dedupeWebhookTargets collapses duplicate dispatch targets. Two targets are
// considered duplicates when (URL, Secret, Match.Canonical()) match — same
// URL with a different scope is intentionally a distinct target so a single
// URL can subscribe to two different scoped views in the same delivery.
func dedupeWebhookTargets(targets []webhookTarget) []webhookTarget {
	uniqueTargets := make([]webhookTarget, 0, len(targets))
	seen := make(map[string]struct{}, len(targets))

	for _, target := range targets {
		if target.URL == "" {
			continue
		}

		key := target.URL + "\x00" + target.Secret + "\x00" + target.Match.Canonical()
		if _, ok := seen[key]; ok {
			continue
		}

		seen[key] = struct{}{}
		uniqueTargets = append(uniqueTargets, target)
	}

	return uniqueTargets
}

// buildContextLinks builds dashboard URLs for the entities referenced in the
// event. Returns nil when AppURL is unconfigured or no link could be produced.
func (h *WebhookSignalLifecycleHook) buildContextLinks(event signal.SignalPhaseEvent, step *workflowStepRef) *contextLinks {
	if h.appURL == "" || event.OrgID == "" {
		return nil
	}

	links := &contextLinks{
		Org: h.dashboardURL(event.OrgID),
	}

	// Owners of install workflows are installs; resolve install id from the
	// owner block when applicable.
	var installID string
	if event.OwnerType == "installs" && event.OwnerID != "" {
		installID = event.OwnerID
	} else if event.InstallID != nil && *event.InstallID != "" {
		installID = *event.InstallID
	}

	if installID != "" {
		links.Install = h.dashboardURL(event.OrgID, "installs", installID)
		if event.WorkflowID != "" {
			links.Workflow = h.dashboardURL(event.OrgID, "installs", installID, "workflows", event.WorkflowID)
		}
		if step != nil {
			if step.SandboxID != "" {
				links.Sandbox = h.dashboardURL(event.OrgID, "installs", installID, "sandbox")
			}
			if step.ComponentID != "" {
				links.Component = h.dashboardURL(event.OrgID, "installs", installID, "components", step.ComponentID)
			}
		}
	}

	if links.Org == "" && links.Install == "" && links.Workflow == "" && links.Sandbox == "" && links.Component == "" {
		return nil
	}
	return links
}

// dashboardURL joins the configured AppURL with the given path pieces. Returns
// an empty string if AppURL is unset or the join fails.
func (h *WebhookSignalLifecycleHook) dashboardURL(pieces ...string) string {
	if h.appURL == "" {
		return ""
	}
	link, err := url.JoinPath(h.appURL, pieces...)
	if err != nil {
		return ""
	}
	return link
}

func normalizeWebhookURLs(webhookURLs []string) []string {
	clean := make([]string, 0, len(webhookURLs))
	for _, webhookURL := range webhookURLs {
		trimmed := strings.TrimSpace(webhookURL)
		if trimmed == "" {
			continue
		}

		clean = append(clean, trimmed)
	}

	return clean
}

func webhookHost(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsed.Host
}
