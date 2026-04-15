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
	return "signal_lifecycle_webhook"
}

func (h *WebhookSignalLifecycleHook) Supports(event signal.SignalPhaseEvent) bool {
	if len(h.webhookURLs) == 0 && h.db == nil {
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

func (h *WebhookSignalLifecycleHook) BeforePhase(ctx context.Context, event signal.SignalPhaseEvent) (signal.BeforePhaseDecision, error) {
	if err := h.publish(ctx, event, nil, "before"); err != nil {
		h.l.Debug("failed to publish pre-phase signal lifecycle webhook", zap.Error(err))
	}

	return signal.AllowPhaseDecision(), nil
}

func (h *WebhookSignalLifecycleHook) AfterPhase(ctx context.Context, event signal.SignalPhaseEvent, outcome signal.SignalPhaseOutcome) error {
	h.l.Debug("webhook after-phase called",
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("phase", string(event.Phase)),
		zap.String("operation", event.Operation),
		zap.String("status", string(outcome.Status)),
	)
	return h.publish(ctx, event, &outcome, "after")
}

// CloudEvents v1.0 envelope types.

type cloudEvent struct {
	SpecVersion     string `json:"specversion"`
	ID              string `json:"id"`
	Type            string `json:"type"`
	Source          string `json:"source"`
	Time            string `json:"time"`
	Subject         string `json:"subject"`
	DataContentType string `json:"datacontenttype"`
	NuonOrgID       string `json:"nuonorgid,omitempty"`
	NuonOperation   string `json:"nuonoperation,omitempty"`
	NuonTransition  string `json:"nuontransition"`
	Data            any    `json:"data"`
}

type signalLifecycleEventData struct {
	Signal    signalIdentity `json:"signal"`
	Phase     string         `json:"phase"`
	Operation string         `json:"operation"`
	Context   signalContext  `json:"context"`
	Outcome   *signalOutcome `json:"outcome"`
}

type signalIdentity struct {
	ID       string  `json:"id"`
	QueueID  string  `json:"queue_id"`
	Type     string  `json:"type"`
	ParentID *string `json:"parent_id"`
}

type signalContext struct {
	OrgID       string  `json:"org_id"`
	AppID       *string `json:"app_id"`
	InstallID   *string `json:"install_id"`
	ComponentID *string `json:"component_id"`
	SandboxID   *string `json:"sandbox_id"`
	RunnerID    *string `json:"runner_id"`
}

type signalOutcome struct {
	Status     string         `json:"status"`
	Error      string         `json:"error,omitempty"`
	DurationMs int64          `json:"duration_ms"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type webhookTarget struct {
	URL    string
	Secret string
}

func (h *WebhookSignalLifecycleHook) publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, transition string) error {
	var dataOutcome *signalOutcome
	if outcome != nil && transition == "after" {
		dataOutcome = &signalOutcome{
			Status:     string(outcome.Status),
			Error:      outcome.ErrMessage,
			DurationMs: outcome.Duration.Milliseconds(),
			Metadata:   outcome.Metadata,
		}
	}

	ce := cloudEvent{
		SpecVersion:     "1.0",
		ID:              uuid.New().String(),
		Type:            "com.nuon.signal.lifecycle.v1",
		Source:          "//nuon.co/ctl-api",
		Time:            time.Now().UTC().Format(time.RFC3339),
		Subject:         event.QueueSignalID,
		DataContentType: "application/json",
		NuonOrgID:       event.OrgID,
		NuonOperation:   event.Operation,
		NuonTransition:  transition,
		Data: signalLifecycleEventData{
			Signal: signalIdentity{
				ID:      event.QueueSignalID,
				QueueID: event.QueueID,
				Type:    string(event.SignalType),
			},
			Phase:     string(event.Phase),
			Operation: event.Operation,
			Context: signalContext{
				OrgID:       event.OrgID,
				InstallID:   event.InstallID,
				ComponentID: event.ComponentID,
			},
			Outcome: dataOutcome,
		},
	}

	payloadJSON, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("unable to marshal signal lifecycle webhook payload: %w", err)
	}

	logger := h.l.With(
		zap.String("hook", h.Name()),
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("phase", string(event.Phase)),
		zap.String("transition", transition),
	)

	targets := make([]webhookTarget, 0, len(h.webhookURLs))
	for _, webhookURL := range h.webhookURLs {
		targets = append(targets, webhookTarget{URL: webhookURL})
	}

	dynamicTargets, err := h.listOrgWebhookTargets(ctx, event.OrgID)
	if err != nil {
		logger.Warn("failed to resolve org signal lifecycle webhooks", zap.Error(err))
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
			logger.Warn("failed to deliver signal lifecycle webhook",
				zap.String("webhook_host", webhookHost(target.URL)),
				zap.Error(err))
		}
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}

	logger.Debug("delivered signal lifecycle webhook")
	return nil
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
		return nil, fmt.Errorf("unable to list org signal lifecycle webhooks: %w", err)
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
