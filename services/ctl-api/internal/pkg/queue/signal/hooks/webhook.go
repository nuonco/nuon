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
	"go.opentelemetry.io/otel/metric"
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
	cloudEventTypeComponentHealth           = "com.nuon.component.health.v1"
	cloudEventTypeInstallHealth             = "com.nuon.install.health.v1"
	cloudEventTypeInstallSync               = "com.nuon.app.install_sync.v1"
	cloudEventTypeInstallConfigSync         = "com.nuon.install.config_sync.v1"
	cloudEventTypeLabelAdded                = "com.nuon.install.label_added.v1"
	cloudEventTypeAppBranchChanged          = "com.nuon.install.app_branch_changed.v1"

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
	kindLabelAdded           = "label_added"
	kindAppBranchChanged     = "app_branch_changed"
)

const (
	statusStarted   = "started"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
	statusCanceled  = "cancelled"
)

const (
	transitionStarted   = "started"
	transitionSucceeded = "succeeded"
	transitionFailed    = "failed"
	transitionCanceled  = "cancelled"

	transitionRequested = "requested"
	transitionApproved  = "approved"
	transitionRejected  = "rejected"
	transitionUnhealthy = "unhealthy"

	transitionAwaitingRetry = "awaiting_retry"

	transitionRecovered = "recovered"
)

const (
	signalTypeExecuteWorkflow              signal.SignalType = "execute-workflow"
	signalTypeExecuteWorkflowStep          signal.SignalType = "execute-workflow-step"
	signalTypeWorkflowStepApprovalRequest  signal.SignalType = "workflow-step-approval-request"
	signalTypeWorkflowStepApprovalResponse signal.SignalType = "workflow-step-approval-response"
	signalTypeDriftDetected                signal.SignalType = "drift-detected"
	signalTypeWorkflowStepAwaitingRetry    signal.SignalType = "workflow-step-awaiting-retry"

	signalTypeStackRun        signal.SignalType = "stack-run"
	signalTypeRoleChange      signal.SignalType = "role-change"
	signalTypeInputsUpdated   signal.SignalType = "inputs-updated"
	signalTypeAppConfigSynced signal.SignalType = "app-config-synced"
	signalTypeUpdateAppConfig signal.SignalType = "update-app-config"
	signalTypeRunnerUnhealthy signal.SignalType = "runner-unhealthy"

	signalTypeComponentUnhealthy signal.SignalType = "component-unhealthy"
	signalTypeComponentRecovered signal.SignalType = "component-recovered"
	signalTypeInstallDegraded    signal.SignalType = "install-degraded"

	signalTypeSyncInstalls      signal.SignalType = "sync-installs"
	signalTypeInstallConfigSync signal.SignalType = "install-config-sync"
	signalTypeLabelAdded        signal.SignalType = "label-added"
	signalTypeAppBranchChanged  signal.SignalType = "app-branch-changed"
)

const approvalPlanExcerptMaxBytes = 8 * 1024

const orgNameCacheTTL = 10 * time.Minute

type Params struct {
	fx.In

	Cfg           *internal.Config     `optional:"true"`
	L             *zap.Logger          `optional:"true"`
	DB            *gorm.DB             `name:"psql" optional:"true"`
	MW            metrics.Writer       `optional:"true"`
	MeterProvider metric.MeterProvider `optional:"true"`
}

type WebhookSignalLifecycleHook struct {
	l               *zap.Logger
	httpClient      *http.Client
	webhookURLs     []string
	db              *gorm.DB
	appURL          string
	publicAPIURL    string
	mw              metrics.Writer
	deliveryMetrics *deliveryMetrics
	blobReadEnabled bool

	workflowCreatorCache sync.Map

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
		deliveryMetrics: newDeliveryMetrics(params.MeterProvider),
		blobReadEnabled: blobReadEnabled,
	}
}

func (h *WebhookSignalLifecycleHook) metricNamespace(ctx context.Context) string {
	info := activity.GetInfo(ctx)
	return info.WorkflowNamespace
}

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
		signalTypeInstallConfigSync,
		signalTypeLabelAdded,
		signalTypeAppBranchChanged:
		return true
	default:
		return false
	}
}

func (h *WebhookSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}

	if suppressesStartedEvent(event.SignalType) {
		return signal.AllowPhaseDecision(), nil
	}

	if err := h.publish(ctx, event, nil); err != nil {
		h.l.Debug("failed to publish workflow lifecycle started webhook", zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

func isApprovalSignalType(t signal.SignalType) bool {
	return t == signalTypeWorkflowStepApprovalRequest ||
		t == signalTypeWorkflowStepApprovalResponse
}

func isNotificationOnlySignalType(t signal.SignalType) bool {
	switch t {
	case signalTypeDriftDetected, signalTypeStackRun, signalTypeRoleChange, signalTypeInputsUpdated, signalTypeAppConfigSynced, signalTypeUpdateAppConfig,
		signalTypeRunnerUnhealthy,
		signalTypeComponentUnhealthy, signalTypeComponentRecovered, signalTypeInstallDegraded,
		signalTypeSyncInstalls, signalTypeInstallConfigSync, signalTypeLabelAdded, signalTypeAppBranchChanged:
		return true
	}
	return false
}

func isComponentHealthSignalType(t signal.SignalType) bool {
	return t == signalTypeComponentUnhealthy ||
		t == signalTypeComponentRecovered ||
		t == signalTypeInstallDegraded
}

// why: suppressesStartedEvent reports whether a signal type's synthetic "started"
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
	if event.Phase == signal.SignalPhaseValidate {
		return nil
	}

	suppress, err := resolveFlowCompletionOutcome(ctx, workflowStatusFromDB(h.db), event, &outcome)
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

type workflowStatusLookup func(ctx context.Context, workflowID string) (app.CompositeStatus, error)

func workflowStatusFromDB(db *gorm.DB) workflowStatusLookup {
	if db == nil {
		return nil
	}
	return func(ctx context.Context, workflowID string) (app.CompositeStatus, error) {
		var flw workflowStatusRow
		err := retryDBRead(ctx, func() error {
			return db.WithContext(ctx).
				Select("status").
				Where(workflowStatusRow{ID: workflowID}).
				First(&flw).Error
		})
		return flw.Status, err
	}
}

// why: resolveFlowCompletionOutcome reconciles an execute-workflow completion
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
func resolveFlowCompletionOutcome(ctx context.Context, lookup workflowStatusLookup, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (bool, error) {
	if lookup == nil ||
		event.SignalType != signalTypeExecuteWorkflow ||
		event.Phase != signal.SignalPhaseExecute ||
		event.WorkflowID == "" ||
		outcome.Status != signal.SignalStatusSuccess {
		return false, nil
	}

	status, err := lookup(ctx, event.WorkflowID)
	if err != nil {
		return false, fmt.Errorf("unable to load workflow status for lifecycle completion: %w", err)
	}

	switch status.Status {
	case app.StatusFailedPendingRetry:
		return true, nil
	case app.StatusError:
		outcome.Status = signal.SignalStatusError
		outcome.ErrMessage = status.StatusHumanDescription
		if outcome.ErrMessage == "" {
			outcome.ErrMessage = "workflow failed"
		}
	case app.StatusCancelled:
		outcome.Status = signal.SignalStatusCancelled
		outcome.ErrMessage = status.StatusHumanDescription
		if outcome.ErrMessage == "" {
			outcome.ErrMessage = "workflow cancelled"
		}
	}

	return false, nil
}

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

	Interests []string `json:"interests,omitempty"`

	Data lifecycleEventData `json:"data"`
}

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
	ID             string    `json:"id"`
	Type           string    `json:"type,omitempty"`
	OwnerID        string    `json:"owner_id,omitempty"`
	OwnerType      string    `json:"owner_type,omitempty"`
	OwnerName      string    `json:"owner_name,omitempty"`
	CreatedByEmail string    `json:"created_by_email,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	RunbookName    string    `json:"runbook_name,omitempty"`
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
	ActionName string `json:"action_name,omitempty"`
}

type lifecycleOutcome struct {
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
	DurationMs int64  `json:"duration_ms,omitempty"`
}

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

type webhookTarget struct {
	URL        string
	Secret     string
	ConfigOnly bool
	Interests  interests.Interests
	Match      *labels.SubscriptionMatch
}

func (h *WebhookSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
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

	targets := make([]webhookTarget, 0, len(h.webhookURLs))
	for _, webhookURL := range h.webhookURLs {
		targets = append(targets, webhookTarget{
			URL:        webhookURL,
			ConfigOnly: true,
		})
	}

	dynamicTargets, err := h.listOrgWebhookTargets(ctx, event.OrgID)
	if err != nil {
		// TODO: Surface lookup failures in hook metrics without preventing delivery to static targets.
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
	case kindLabelAdded:
		ceType = cloudEventTypeLabelAdded
	case kindAppBranchChanged:
		ceType = cloudEventTypeAppBranchChanged
	}
	if event.SignalType == signalTypeWorkflowStepAwaitingRetry {
		ceType = cloudEventTypeWorkflowStepAwaitingRetry
	}

	subject := buildSubject(event, data)

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

	matchTargets := EventTargetsFromEvent(ctx, h.db, event, data)

	labelLoader := newLabelLoader(h.db)

	var sendErrs []error
	for _, target := range targets {
		if !target.ConfigOnly && target.Match != nil {
			if err := labelLoader.load(ctx, &matchTargets); err != nil {
				logger.Warn("failed to load event labels for match",
					zap.Error(err))
			}
			if !target.Match.Matches(matchTargets) {
				continue
			}
		}

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
	case signalTypeLabelAdded:
		return h.buildLabelAddedEventData(event, outcome)
	case signalTypeAppBranchChanged:
		data, ok := h.buildLabelAddedEventData(event, outcome)
		data.Kind = kindAppBranchChanged
		return data, ok
	}

	if event.WorkflowID == "" {
		return lifecycleEventData{}, false
	}

	if isApprovalSignalType(event.SignalType) {
		return h.buildApprovalEventData(ctx, event, outcome)
	}

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
		OrgName:    event.OrgName,
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

	if event.SignalType == signalTypeWorkflowStepAwaitingRetry {
		data.Metadata = event.Metadata
		if data.Outcome != nil {
			data.Outcome.Status = statusFailed
			if errMsg, _ := event.Metadata["error"].(string); errMsg != "" {
				data.Outcome.Error = errMsg
			}
		}
	}

	if event.StepID != "" && (kind == kindWorkflowStep || event.SignalType == signalTypeDriftDetected) {
		stepRef, installName, emit := h.enrichStep(ctx, event.StepID)
		if !emit {
			return lifecycleEventData{}, false
		}
		if isAwaitRunnerHealthyStep(event, stepRef) {
			return lifecycleEventData{}, false
		}
		data.Step = stepRef
		if installName != "" && data.Workflow.OwnerType == "installs" && data.Workflow.OwnerName == "" {
			data.Workflow.OwnerName = installName
		}
	}

	data.Parent = h.lookupParent(ctx, event.WorkflowID)

	data.Links = h.buildContextLinks(event, data.Step)

	return data, true
}

func isAwaitRunnerHealthyStep(event signal.SignalPhaseEvent, step *workflowStepRef) bool {
	return event.StepName == "runner healthy" ||
		step != nil && step.TargetType == string(app.WorkflowStepTargetTypeRunners)
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

func (h *WebhookSignalLifecycleHook) buildLabelAddedEventData(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	data := lifecycleEventData{
		Kind:       kindLabelAdded,
		Transition: mapTransition(event, outcome),
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

	var linkStep *workflowStepRef
	if event.ComponentID != nil && *event.ComponentID != "" {
		linkStep = &workflowStepRef{ComponentID: *event.ComponentID}
	}
	data.Links = h.buildContextLinks(event, linkStep)

	return data, true
}

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

func (h *WebhookSignalLifecycleHook) buildApprovalEventData(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	if event.StepID == "" || h.db == nil {
		return lifecycleEventData{}, false
	}
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
		data.Links.Approval = data.Links.Workflow
		data.Links.RespondAPI = h.respondAPIURL(event.WorkflowID, event.StepID, approval.ID)
	}

	return data, true
}

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

type approvalResponseRow struct {
	Type        app.WorkflowStepResponseType
	RespondedBy string
}

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

	// why: Both singular and plural target type strings exist in the codebase
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

type deployTargetMeta struct {
	ComponentID   string
	ComponentName string
	InstallName   string
}

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

type sandboxRunTargetMeta struct {
	SandboxID   string
	InstallName string
}

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

func (h *WebhookSignalLifecycleHook) lookupParent(ctx context.Context, workflowID string) *parentRef {
	if h.db == nil || workflowID == "" {
		return nil
	}

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

func (h *WebhookSignalLifecycleHook) sendWebhook(ctx context.Context, target webhookTarget, payloadJSON []byte) (retErr error) {
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

	started := time.Now()
	defer func() {
		h.deliveryMetrics.record(ctx, deliveryChannelWebhook, deliveryOperationPost, started, retErr)
	}()

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

		// why: Treat an unconfigured (zero-value) Interests as "all events".
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

func (h *WebhookSignalLifecycleHook) buildContextLinks(event signal.SignalPhaseEvent, step *workflowStepRef) *contextLinks {
	if h.appURL == "" || event.OrgID == "" {
		return nil
	}

	links := &contextLinks{
		Org: h.dashboardURL(event.OrgID),
	}

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
