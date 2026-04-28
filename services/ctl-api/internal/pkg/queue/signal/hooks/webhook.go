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

// userFacingOperations is the set of operations that should be exposed
// via webhooks. Internal-only operations are filtered out.
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

// Stage values surfaced to webhook consumers. Empty stage means a
// single-phase operation.
const (
	stagePlan  = "plan"
	stageApply = "apply"
)

// Status values surfaced to webhook consumers in the *.finished events.
const (
	statusStarted   = "started"
	statusSucceeded = "succeeded"
	statusFailed    = "failed"
	statusCanceled  = "canceled"
)

// Failure reasons surfaced when an event's status is "failed".
const (
	failureReasonValidationFailed = "validation_failed"
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
		httpClient: &http.Client{
			Timeout: timeout,
		},
		webhookURLs: normalizeWebhookURLs(webhookURLs),
		db:          params.DB,
	}
}

func (h *WebhookSignalLifecycleHook) Name() string {
	return "operation_lifecycle_webhook"
}

func (h *WebhookSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if len(h.webhookURLs) == 0 && h.db == nil {
		return false
	}

	if event.Operation == "" {
		return false
	}

	if _, ok := userFacingOperations[event.Operation]; !ok {
		return false
	}

	// Validate phase events are only surfaced when validation FAILS.
	// Successful validates remain silent — we only want users to see them
	// when they translate into a terminal *.finished event.
	// The Supports() filter is permissive here; the publish path decides
	// whether to actually emit anything.
	return true
}

func (h *WebhookSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	// Only emit *.started events for the execute phase. Validate/cancel
	// before-phase callbacks are never user-facing.
	if event.Phase != signal.SignalPhaseExecute {
		return signal.AllowPhaseDecision(), nil
	}

	if err := h.publish(ctx, event, nil); err != nil {
		h.l.Debug("failed to publish operation lifecycle started webhook", zap.Error(err))
	}
	return signal.AllowPhaseDecision(), nil
}

func (h *WebhookSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	// Suppress noise for successful validates — we only emit on validate failure.
	if event.Phase == signal.SignalPhaseValidate && outcome.Status == signal.SignalStatusSuccess {
		return nil
	}

	h.l.Debug("webhook after-phase called",
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("phase", string(event.Phase)),
		zap.String("operation", event.Operation),
		zap.String("stage", event.Stage),
		zap.String("status", string(outcome.Status)),
	)

	return h.publish(ctx, event, &outcome)
}

// CloudEvents v1.0 envelope. We use CloudEvent extension attributes
// (lowercased, alphanumeric) to expose org/operation/stage/status without
// leaking internal signal/queue/phase terminology into the wire format.
type cloudEvent struct {
	SpecVersion     string `json:"specversion"`
	ID              string `json:"id"`
	Type            string `json:"type"`
	Source          string `json:"source"`
	Time            string `json:"time"`
	Subject         string `json:"subject"`
	DataContentType string `json:"datacontenttype"`

	NuonOrgID     string `json:"nuonorgid,omitempty"`
	NuonOperation string `json:"nuonoperation,omitempty"`
	NuonStage     string `json:"nuonstage,omitempty"`
	NuonStatus    string `json:"nuonstatus,omitempty"`

	Data operationLifecycleEventData `json:"data"`
}

type operationLifecycleEventData struct {
	Event         string           `json:"event"`
	Operation     string           `json:"operation"`
	Stage         string           `json:"stage,omitempty"`
	Status        string           `json:"status"`
	FailureReason string           `json:"failure_reason,omitempty"`
	Error         string           `json:"error,omitempty"`
	DurationMs    int64            `json:"duration_ms,omitempty"`
	Context       operationContext `json:"context"`
	Metadata      map[string]any   `json:"metadata,omitempty"`
}

type operationContext struct {
	OrgID       string  `json:"org_id"`
	InstallID   *string `json:"install_id,omitempty"`
	ComponentID *string `json:"component_id,omitempty"`
	SandboxID   *string `json:"sandbox_id,omitempty"`
}

type webhookTarget struct {
	URL    string
	Secret string
}

// publish builds and emits the CloudEvent for an operation lifecycle
// transition. When outcome is nil this is a *.started event; otherwise it is
// a *.finished event carrying status/error/duration.
func (h *WebhookSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) error {
	data, ok := h.buildEventData(event, outcome)
	if !ok {
		return nil
	}

	subject := buildSubject(event, data.Event)

	ce := cloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Type:            "com.nuon.operation.lifecycle.v1",
		Source:          "//nuon.co/ctl-api",
		Time:            time.Now().UTC().Format(time.RFC3339),
		Subject:         subject,
		DataContentType: "application/json",
		NuonOrgID:       event.OrgID,
		NuonOperation:   data.Operation,
		NuonStage:       data.Stage,
		NuonStatus:      data.Status,
		Data:            data,
	}

	payloadJSON, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("unable to marshal operation lifecycle webhook payload: %w", err)
	}

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("operation", data.Operation),
		zap.String("stage", data.Stage),
		zap.String("status", data.Status),
		zap.String("event", data.Event),
	)

	targets := make([]webhookTarget, 0, len(h.webhookURLs))
	for _, webhookURL := range h.webhookURLs {
		targets = append(targets, webhookTarget{URL: webhookURL})
	}

	dynamicTargets, err := h.listOrgWebhookTargets(ctx, event.OrgID)
	if err != nil {
		logger.Warn("failed to resolve org operation lifecycle webhooks", zap.Error(err))
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
			logger.Warn("failed to deliver operation lifecycle webhook",
				zap.String("webhook_host", webhookHost(target.URL)),
				zap.Error(err))
		}
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}

	logger.Debug("delivered operation lifecycle webhook")
	return nil
}

// buildEventData translates an internal SignalPhaseEvent + outcome into the
// user-facing event payload. Returns ok=false when the input doesn't map to a
// user-facing transition (e.g. a phase we deliberately suppress).
func (h *WebhookSignalLifecycleHook) buildEventData(event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome) (operationLifecycleEventData, bool) {
	data := operationLifecycleEventData{
		Operation: event.Operation,
		Stage:     event.Stage,
		Context: operationContext{
			OrgID:       event.OrgID,
			InstallID:   event.InstallID,
			ComponentID: event.ComponentID,
			SandboxID:   event.SandboxID,
		},
	}

	stagePrefix := stageEventPrefix(event.Stage)

	if outcome == nil {
		// before-phase: started transition.
		data.Event = stagePrefix + ".started"
		data.Status = statusStarted
		return data, true
	}

	// after-phase: finished transition.
	data.Event = stagePrefix + ".finished"
	data.Status = mapStatus(outcome.Status)
	if outcome.Status != signal.SignalStatusSuccess {
		data.Error = outcome.ErrMessage
	}
	if outcome.Duration > 0 {
		data.DurationMs = outcome.Duration.Milliseconds()
	}
	if len(outcome.Metadata) > 0 {
		data.Metadata = outcome.Metadata
	}

	// Validate-phase failures translate into a *.finished event for the
	// owning stage, with failure_reason="validation_failed". Successful
	// validates are filtered out earlier in AfterPhase.
	if event.Phase == signal.SignalPhaseValidate {
		data.Status = statusFailed
		data.FailureReason = failureReasonValidationFailed
	}

	// Cancel phase always reports canceled status regardless of outcome.
	if event.Phase == signal.SignalPhaseCancel {
		data.Status = statusCanceled
	}

	return data, true
}

// stageEventPrefix returns the user-facing event family ("plan", "apply", or
// "operation") for a given stage value.
func stageEventPrefix(stage string) string {
	switch stage {
	case stagePlan:
		return "plan"
	case stageApply:
		return "apply"
	default:
		return "operation"
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

// buildSubject returns a stable, non-signal-leaking identifier for the
// CloudEvent's `subject`. It composes the org id, operation, stage, and event
// family so consumers can correlate started/finished pairs without exposing
// any internal signal/queue identifiers in the wire format.
func buildSubject(event signal.SignalPhaseEvent, eventName string) string {
	parts := []string{}
	if event.OrgID != "" {
		parts = append(parts, event.OrgID)
	}
	if event.Operation != "" {
		parts = append(parts, event.Operation)
	}
	if event.Stage != "" {
		parts = append(parts, event.Stage)
	}
	parts = append(parts, eventName)
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
		return nil, fmt.Errorf("unable to list org operation lifecycle webhooks: %w", err)
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
