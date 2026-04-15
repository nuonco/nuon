package webhook

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
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

// Target describes a single webhook destination.
type Target struct {
	URL    string
	Secret string
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

// Deliverer handles building CloudEvents payloads, resolving targets, and
// delivering webhook HTTP requests.
type Deliverer struct {
	l          *zap.Logger
	httpClient *http.Client
	staticURLs []string
	db         *gorm.DB
}

// NewDeliverer creates a Deliverer with the given configuration.
func NewDeliverer(l *zap.Logger, httpClient *http.Client, staticURLs []string, db *gorm.DB) *Deliverer {
	return &Deliverer{
		l:          l,
		httpClient: httpClient,
		staticURLs: NormalizeURLs(staticURLs),
		db:         db,
	}
}

// Publish builds a CloudEvents payload from the signal event and delivers it
// to all resolved webhook targets.
func (d *Deliverer) Publish(ctx context.Context, event signal.SignalPhaseEvent, outcome *signal.SignalPhaseOutcome, transition string) error {
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

	logger := d.l.With(
		zap.String("queue_signal_id", event.QueueSignalID),
		zap.String("signal_type", string(event.SignalType)),
		zap.String("phase", string(event.Phase)),
		zap.String("transition", transition),
	)

	targets := make([]Target, 0, len(d.staticURLs))
	for _, webhookURL := range d.staticURLs {
		targets = append(targets, Target{URL: webhookURL})
	}

	dynamicTargets, err := d.listOrgTargets(ctx, event.OrgID)
	if err != nil {
		logger.Warn("failed to resolve org signal lifecycle webhooks", zap.Error(err))
	}

	targets = append(targets, dynamicTargets...)
	targets = DedupeTargets(targets)
	if len(targets) == 0 {
		return nil
	}

	logger = logger.With(zap.Int("webhook_count", len(targets)))

	var sendErrs []error
	for _, target := range targets {
		if err := d.send(ctx, target, payloadJSON); err != nil {
			sendErrs = append(sendErrs, err)
			logger.Warn("failed to deliver signal lifecycle webhook",
				zap.String("webhook_host", hostFromURL(target.URL)),
				zap.Error(err))
		}
	}

	if len(sendErrs) > 0 {
		return errors.Join(sendErrs...)
	}

	logger.Debug("delivered signal lifecycle webhook")
	return nil
}

func (d *Deliverer) send(ctx context.Context, target Target, payloadJSON []byte) error {
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

	resp, err := d.httpClient.Do(req)
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

func (d *Deliverer) listOrgTargets(ctx context.Context, orgID string) ([]Target, error) {
	if d.db == nil || orgID == "" {
		return nil, nil
	}

	var webhooks []app.Webhook
	if err := d.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("unable to list org signal lifecycle webhooks: %w", err)
	}

	targets := make([]Target, 0, len(webhooks))
	for _, wh := range webhooks {
		trimmedURL := strings.TrimSpace(wh.WebhookURL)
		if trimmedURL == "" {
			continue
		}

		targets = append(targets, Target{
			URL:    trimmedURL,
			Secret: strings.TrimSpace(wh.WebhookSecret),
		})
	}

	return targets, nil
}

// DedupeTargets removes duplicate webhook targets by URL+Secret pair.
func DedupeTargets(targets []Target) []Target {
	uniqueTargets := make([]Target, 0, len(targets))
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

// NormalizeURLs trims whitespace and drops empty entries.
func NormalizeURLs(webhookURLs []string) []string {
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

func hostFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	return parsed.Host
}
