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
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Public webhook primitives. Consumers reason about exactly two things:
// the workflow lifecycle and the workflow step lifecycle. Operation taxonomy
// (component-deploy, sandbox-provision, etc.), multi-phase concepts (plan/apply),
// and inner signal type names are deliberately NOT exposed.
const (
	cloudEventTypeWorkflow     = "com.nuon.workflow.lifecycle.v1"
	cloudEventTypeWorkflowStep = "com.nuon.workflow_step.lifecycle.v1"

	kindWorkflow     = "workflow"
	kindWorkflowStep = "workflow_step"
)

// Status values surfaced to webhook consumers in the *.lifecycle events.
const (
	statusStarted   = "started"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
	statusCanceled  = "cancelled"
)

// Transition values surfaced in `data.transition`. Mirrors statuses with the
// same name; the dedicated field exists so consumers can switch on transitions
// without reading the (potentially redundant) `status` field.
const (
	transitionStarted   = "started"
	transitionSucceeded = "succeeded"
	transitionFailed    = "failed"
	transitionCanceled  = "cancelled"
)

// signalTypeExecuteWorkflow matches the SignalType produced by
// services/ctl-api/internal/pkg/flow/signals/executeflow. Duplicated as a string
// constant to avoid importing the flow package and producing an import cycle.
const (
	signalTypeExecuteWorkflow     signal.SignalType = "execute-workflow"
	signalTypeExecuteWorkflowStep signal.SignalType = "execute-workflow-step"
)

type Params struct {
	fx.In

	Cfg *internal.Config `optional:"true"`
	L   *zap.Logger      `optional:"true"`
	DB  *gorm.DB         `name:"psql" optional:"true"`
}

type WebhookSignalLifecycleHook struct {
	l           *zap.Logger
	httpClient  *http.Client
	webhookURLs []string
	db          *gorm.DB
	appURL      string
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
	if params.Cfg != nil {
		webhookURLs = params.Cfg.WebhookURLs
		appURL = strings.TrimSpace(params.Cfg.AppURL)
	}

	return &WebhookSignalLifecycleHook{
		l: logger,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		webhookURLs: normalizeWebhookURLs(webhookURLs),
		db:          params.DB,
		appURL:      appURL,
	}
}

func (h *WebhookSignalLifecycleHook) Name() string {
	return "workflow_lifecycle_webhook"
}

// Supports limits this hook to the two public lifecycle primitives:
// execute-workflow (workflow lifecycle) and execute-workflow-step (step
// lifecycle). Inner-signal events (plan/apply, component-deploy, etc.) are
// deliberately ignored — consumers should reason in terms of workflow + step.
func (h *WebhookSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if len(h.webhookURLs) == 0 && h.db == nil {
		return false
	}

	switch event.SignalType {
	case signalTypeExecuteWorkflow, signalTypeExecuteWorkflowStep:
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

	if err := h.publish(ctx, event, nil); err != nil {
		h.l.Debug("failed to publish workflow lifecycle started webhook", zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

func (h *WebhookSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	// Validation phases never produce a public event for these primitives —
	// validation failures of the workflow / step wrappers surface as a failed
	// execute outcome immediately afterward.
	if event.Phase == signal.SignalPhaseValidate {
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

	Data lifecycleEventData `json:"data"`
}

// lifecycleEventData is the public webhook payload. It exposes only two
// primitives — workflow and (optionally) step — alongside the transition,
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
	Links    *contextLinks     `json:"links,omitempty"`
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

type contextLinks struct {
	Org       string `json:"org,omitempty"`
	Install   string `json:"install,omitempty"`
	Workflow  string `json:"workflow,omitempty"`
	Sandbox   string `json:"sandbox,omitempty"`
	Component string `json:"component,omitempty"`
}

type webhookTarget struct {
	URL    string
	Secret string
}

func (h *WebhookSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	data, ok := h.buildEventData(ctx, event, outcome)
	if !ok {
		return nil
	}

	ceType := cloudEventTypeWorkflow
	if data.Kind == kindWorkflowStep {
		ceType = cloudEventTypeWorkflowStep
	}

	subject := buildSubject(event, data)

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
		Data:            data,
	}

	payloadJSON, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("unable to marshal workflow lifecycle webhook payload: %w", err)
	}

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("kind", data.Kind),
		zap.String("transition", data.Transition),
		zap.String("workflow_id", data.Workflow.ID),
	)

	targets := make([]webhookTarget, 0, len(h.webhookURLs))
	for _, webhookURL := range h.webhookURLs {
		targets = append(targets, webhookTarget{URL: webhookURL})
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

	logger = logger.With(zap.Int("webhook_count", len(targets)))

	var sendErrs []error
	for _, target := range targets {
		if err := h.sendWebhook(ctx, target, payloadJSON); err != nil {
			sendErrs = append(sendErrs, err)
			logger.Warn("failed to deliver workflow lifecycle webhook",
				zap.String("webhook_host", webhookHost(target.URL)),
				zap.Error(err))
		}
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}

	logger.Debug("delivered workflow lifecycle webhook")
	return nil
}

// buildEventData translates an internal SignalPhaseEvent into the public
// workflow / workflow_step payload. Returns ok=false when there is nothing
// to emit (e.g. missing identifiers).
func (h *WebhookSignalLifecycleHook) buildEventData(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (lifecycleEventData, bool) {
	if event.WorkflowID == "" {
		return lifecycleEventData{}, false
	}

	kind := kindWorkflow
	if event.SignalType == signalTypeExecuteWorkflowStep {
		kind = kindWorkflowStep
	}

	transition := mapTransition(event, outcome)

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

	if outcome != nil {
		data.Outcome = h.buildOutcome(event, outcome)
	}

	if kind == kindWorkflowStep && event.StepID != "" {
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

	var installName string
	switch step.StepTargetType {
	case string(app.WorkflowStepTargetTypeInstallDeploy):
		meta := h.lookupDeployTargetMeta(ctx, step.StepTargetID)
		ref.ComponentID = meta.ComponentID
		ref.ComponentName = meta.ComponentName
		installName = meta.InstallName
	case string(app.WorkflowStepTargetTypeInstallSandboxRun):
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
	if err := h.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("unable to list org workflow lifecycle webhooks: %w", err)
	}

	targets := make([]webhookTarget, 0, len(webhooks))
	for _, webhook := range webhooks {
		trimmedURL := strings.TrimSpace(webhook.WebhookURL)
		if trimmedURL == "" {
			continue
		}

		targets = append(targets, webhookTarget{
			URL:    trimmedURL,
			Secret: strings.TrimSpace(webhook.WebhookSecret),
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

		key := target.URL + "\x00" + target.Secret
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
