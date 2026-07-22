package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queuepkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/slack/signing"
)

const (
	maxEventBody                      = 16 * 1024 * 1024
	maxActiveEventSecrets             = 2
	maxEventSecretMetadata            = 20
	eventSecretRotationGrace          = 24 * time.Hour
	azureEventGridValidationEventType = "Microsoft.EventGrid.SubscriptionValidationEvent"
	slackRequestMaxClockSkew          = 5 * time.Minute
)

var (
	errEventBodyTooLarge   = errors.New("request body too large")
	errTriggerInactive     = errors.New("trigger inactive")
	errEventSecretInactive = errors.New("trigger secret inactive")
	errUnsupportedAuth     = errors.New("trigger auth is not implemented")
	errUnsupportedEnvelope = errors.New("trigger envelope is not implemented")
)

type cloudEvent struct {
	SpecVersion     string          `json:"specversion"`
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	Type            string          `json:"type"`
	Subject         string          `json:"subject,omitempty"`
	Time            *time.Time      `json:"time,omitempty"`
	DataContentType string          `json:"datacontenttype,omitempty"`
	Data            json.RawMessage `json:"data"`
}

type ingressResponse struct {
	EventID   string `json:"event_id"`
	Duplicate bool   `json:"duplicate"`
}

type normalizedEvent struct {
	ID          string
	DedupeID    string
	Source      string
	Type        string
	OccurredAt  *time.Time
	Payload     json.RawMessage
	ContentType string
}

type azureEventGridEvent struct {
	ID        string          `json:"id"`
	EventType string          `json:"eventType"`
	EventTime *time.Time      `json:"eventTime,omitempty"`
	Data      json.RawMessage `json:"data"`
}

type slackEventEnvelope struct {
	Type      string          `json:"type"`
	Challenge string          `json:"challenge,omitempty"`
	EventID   string          `json:"event_id,omitempty"`
	EventTime int64           `json:"event_time,omitempty"`
	Event     json.RawMessage `json:"event,omitempty"`
}

type pubSubPushEnvelope struct {
	Message struct {
		Data        string `json:"data"`
		MessageID   string `json:"messageId"`
		PublishTime string `json:"publishTime"`
	} `json:"message"`
}

func parseCloudEvent(body []byte) (*cloudEvent, error) {
	var event cloudEvent
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&event); err != nil {
		return nil, err
	}
	if decoder.Decode(&struct{}{}) != io.EOF || event.SpecVersion != "1.0" || event.ID == "" || event.Source == "" || event.Type == "" || len(event.Data) == 0 || !json.Valid(event.Data) {
		return nil, errors.New("invalid structured CloudEvent 1.0")
	}
	return &event, nil
}

func decodeEvent(trigger *app.Trigger, headers http.Header, body []byte) (*normalizedEvent, error) {
	if trigger.Preset == "azure-event-grid" {
		return decodeAzureEventGridEvent(body)
	}
	if trigger.Preset == "slack-events" {
		return decodeSlackEvent(body)
	}
	var event *normalizedEvent
	switch trigger.Envelope {
	case app.EventEnvelopeTypeNone:
		if !json.Valid(body) {
			return nil, errors.New("invalid JSON event")
		}
		event = &normalizedEvent{Payload: json.RawMessage(body), ContentType: headers.Get("Content-Type")}
	case app.EventEnvelopeTypeCloudEvents:
		cloudEvent, err := parseCloudEvent(body)
		if err != nil {
			return nil, err
		}
		event = &normalizedEvent{ID: cloudEvent.ID, Source: cloudEvent.Source, Type: cloudEvent.Type, OccurredAt: cloudEvent.Time, Payload: cloudEvent.Data, ContentType: cloudEvent.DataContentType}
	case app.EventEnvelopeTypePubSubPush:
		var push pubSubPushEnvelope
		if err := json.Unmarshal(body, &push); err != nil || push.Message.Data == "" || push.Message.MessageID == "" {
			return nil, errors.New("invalid Pub/Sub push envelope")
		}
		payload, err := base64.StdEncoding.DecodeString(push.Message.Data)
		if err != nil || !json.Valid(payload) {
			return nil, errors.New("invalid Pub/Sub message data")
		}
		var occurredAt *time.Time
		if push.Message.PublishTime != "" {
			parsed, err := time.Parse(time.RFC3339Nano, push.Message.PublishTime)
			if err != nil {
				return nil, errors.New("invalid Pub/Sub publish time")
			}
			occurredAt = &parsed
		}
		event = &normalizedEvent{ID: push.Message.MessageID, OccurredAt: occurredAt, Payload: payload, ContentType: "application/json"}
	case app.EventEnvelopeTypeSNS:
		msg, err := parseSNSMessage(body)
		if err != nil {
			return nil, err
		}
		if msg.Type != "Notification" {
			return nil, nil
		}
		payload := json.RawMessage(msg.Message)
		if !json.Valid(payload) {
			return nil, errors.New("SNS Notification Message must contain JSON")
		}
		occurredAt, err := time.Parse(time.RFC3339Nano, msg.Timestamp)
		if err != nil {
			return nil, errors.New("invalid SNS timestamp")
		}
		event = &normalizedEvent{ID: msg.MessageID, OccurredAt: &occurredAt, Payload: payload, ContentType: "application/json"}
	default:
		return nil, errUnsupportedEnvelope
	}
	if event == nil {
		return nil, nil
	}
	if trigger.Envelope == app.EventEnvelopeTypeCloudEvents {
		event.DedupeID = event.ID
	}
	if err := applyEventSelectors(trigger, headers, event); err != nil {
		return nil, err
	}
	if trigger.Envelope == app.EventEnvelopeTypeNone && event.ID == "" {
		sum := sha256.Sum256(body)
		event.ID = hex.EncodeToString(sum[:])
	}
	if event.DedupeID == "" {
		event.DedupeID = event.ID
	}
	return event, nil
}

func decodeAzureEventGridEvent(body []byte) (*normalizedEvent, error) {
	var events []json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&events); err != nil || decoder.Decode(&struct{}{}) != io.EOF || len(events) != 1 {
		return nil, errors.New("Azure Event Grid request must contain exactly one event")
	}
	var event azureEventGridEvent
	if err := json.Unmarshal(events[0], &event); err != nil || event.ID == "" || event.EventType == "" || len(event.Data) == 0 || !json.Valid(event.Data) || (event.EventType != azureEventGridValidationEventType && event.EventTime == nil) {
		return nil, errors.New("invalid Azure Event Grid event")
	}
	return &normalizedEvent{ID: event.ID, Type: event.EventType, OccurredAt: event.EventTime, Payload: events[0], ContentType: "application/json"}, nil
}

func azureEventGridValidationCode(event *normalizedEvent) (string, error) {
	if event == nil || event.Type != azureEventGridValidationEventType {
		return "", nil
	}
	var payload struct {
		Data struct {
			ValidationCode string `json:"validationCode"`
		} `json:"data"`
	}
	if err := json.Unmarshal(event.Payload, &payload); err != nil || payload.Data.ValidationCode == "" {
		return "", errors.New("Azure Event Grid validation event is missing validationCode")
	}
	return payload.Data.ValidationCode, nil
}

func decodeSlackEvent(body []byte) (*normalizedEvent, error) {
	var envelope slackEventEnvelope
	decoder := json.NewDecoder(bytes.NewReader(body))
	if err := decoder.Decode(&envelope); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return nil, errors.New("invalid Slack event envelope")
	}
	if envelope.Type == "url_verification" {
		if strings.TrimSpace(envelope.Challenge) == "" {
			return nil, errors.New("Slack URL verification request is missing challenge")
		}
		return &normalizedEvent{Type: envelope.Type, Payload: json.RawMessage(body), ContentType: "application/json"}, nil
	}
	if envelope.Type != "event_callback" || strings.TrimSpace(envelope.EventID) == "" || envelope.EventTime <= 0 || len(envelope.Event) == 0 || !json.Valid(envelope.Event) {
		return nil, errors.New("invalid Slack event callback")
	}
	var inner struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(envelope.Event, &inner); err != nil || strings.TrimSpace(inner.Type) == "" {
		return nil, errors.New("Slack event callback is missing event type")
	}
	occurredAt := time.Unix(envelope.EventTime, 0).UTC()
	return &normalizedEvent{ID: envelope.EventID, Type: inner.Type, OccurredAt: &occurredAt, Payload: json.RawMessage(body), ContentType: "application/json"}, nil
}

func slackChallenge(event *normalizedEvent) (string, error) {
	if event == nil || event.Type != "url_verification" {
		return "", nil
	}
	var envelope slackEventEnvelope
	if err := json.Unmarshal(event.Payload, &envelope); err != nil || envelope.Challenge == "" {
		return "", errors.New("Slack URL verification request is missing challenge")
	}
	return envelope.Challenge, nil
}

func validSlackRequestTimestamp(value string, now time.Time) bool {
	seconds, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return false
	}
	drift := now.Sub(time.Unix(seconds, 0))
	return drift <= slackRequestMaxClockSkew && drift >= -slackRequestMaxClockSkew
}

func applyEventSelectors(trigger *app.Trigger, headers http.Header, event *normalizedEvent) error {
	if trigger.IDFrom.Header != "" {
		if value := headers.Get(trigger.IDFrom.Header); value != "" {
			event.ID = value
		}
	}
	if trigger.TypeFrom.Header != "" {
		if value := headers.Get(trigger.TypeFrom.Header); value != "" {
			event.Type = value
		}
	}
	if trigger.IDFrom.Payload == "" && trigger.TypeFrom.Payload == "" {
		return nil
	}
	payload, err := decodeTriggerEventJSON(event.Payload)
	if err != nil {
		return err
	}
	if trigger.IDFrom.Payload != "" {
		event.ID, err = selectEventString(payload, trigger.IDFrom.Payload)
		if err != nil {
			return fmt.Errorf("extract event ID: %w", err)
		}
	}
	if trigger.TypeFrom.Payload != "" {
		event.Type, err = selectEventString(payload, trigger.TypeFrom.Payload)
		if err != nil {
			return fmt.Errorf("extract event type: %w", err)
		}
	}
	return nil
}

func decodeTriggerEventJSON(body []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var payload any
	if err := decoder.Decode(&payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func selectEventString(payload any, pathValue string) (string, error) {
	path, err := eventfilter.ParsePath(pathValue, false)
	if err != nil {
		return "", err
	}
	selected := path.Select(payload)
	if len(selected) != 1 {
		return "", fmt.Errorf("selector matched %d values", len(selected))
	}
	value, ok := selected[0].(string)
	if !ok || value == "" {
		return "", errors.New("selector must match a nonempty string")
	}
	return value, nil
}

func readLimitedBody(body io.Reader) ([]byte, error) {
	limited := io.LimitReader(body, maxEventBody+1)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if len(b) > maxEventBody {
		return nil, errEventBodyTooLarge
	}
	return b, nil
}

// @ID IngestTriggerEvent
// @Summary Ingest an event
// @Tags triggers
// @Accept json
// @Produce json
// @Param ingress_key path string true "Opaque ingress key"
// @Success 202 {object} ingressResponse
// @Router /v1/event-ingress/{ingress_key} [post]
func (s *service) IngestEvent(ctx *gin.Context) {
	var trigger app.Trigger
	err := s.db.WithContext(ctx).Where(app.Trigger{IngressKeyHash: hashIngressKey(ctx.Param("ingress_key")), Status: app.TriggerStatusActive}).First(&trigger).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "trigger not found"})
			return
		}
		ctx.Error(err)
		return
	}
	enabled, err := s.features.OrgHasFeature(ctx, trigger.OrgID, app.OrgFeatureTriggers)
	if err != nil {
		ctx.Error(err)
		return
	}
	if !enabled {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "trigger not found"})
		return
	}
	now := time.Now()
	if authUsesSecret(trigger.AuthType) {
		if err := s.scrubInactiveTriggerSecrets(ctx, s.db, trigger.ID, now); err != nil {
			ctx.Error(err)
			return
		}
		if err := s.loadActiveTriggerSecrets(ctx, &trigger, now); err != nil {
			ctx.Error(err)
			return
		}
	}
	cctx.SetAccountIDGinContext(ctx, trigger.CreatedByID)
	body, err := readLimitedBody(ctx.Request.Body)
	if err != nil {
		if errors.Is(err, errEventBodyTooLarge) {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body exceeds size limit"})
			return
		}
		ctx.Error(err)
		return
	}
	var event *normalizedEvent
	nativeProtocolPreset := trigger.Preset == "azure-event-grid" || trigger.Preset == "slack-events"
	var matched *app.TriggerSecret
	switch trigger.AuthType {
	case app.TriggerAuthTypeNone:
	case app.TriggerAuthTypeHMAC:
		if trigger.Preset == "slack-events" {
			timestamp := ctx.GetHeader(signing.TimestampHeader)
			provided := ctx.GetHeader(signing.SignatureHeader)
			if !validSlackRequestTimestamp(timestamp, now) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Slack request timestamp"})
				return
			}
			for i := range trigger.Secrets {
				secret := &trigger.Secrets[i]
				if activeTriggerSecret(secret, now) && signing.Verify(secret.Secret, timestamp, body, provided) {
					matched = secret
					break
				}
			}
			if matched == nil {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid Slack signature"})
				return
			}
			break
		}
		config := trigger.AuthConfig
		if config.Header == "" {
			config = app.TriggerAuthConfig{Header: "X-Nuon-Signature", Prefix: "v1=", Algorithm: "sha256", Encoding: "hex"}
		}
		signature, err := decodeHMACSignature(ctx.GetHeader(config.Header), config)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature header"})
			return
		}
		for i := range trigger.Secrets {
			secret := &trigger.Secrets[i]
			if !activeTriggerSecret(secret, now) {
				continue
			}
			valid := verifyGenericHMAC(secret.Secret, body, signature, config.Algorithm)
			if valid {
				matched = secret
				break
			}
		}
		if matched == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	case app.TriggerAuthTypeSNSSignature:
		msg, err := parseSNSMessage(body)
		if err != nil || msg.TopicARN != trigger.AuthConfig.TopicARN || s.snsVerifier.verify(ctx, msg) != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid SNS signature"})
			return
		}
		if msg.Type == "SubscriptionConfirmation" {
			if err := s.confirmSNSSubscription(ctx, msg); err != nil {
				ctx.Error(err)
				return
			}
		}
		if msg.Type == "UnsubscribeConfirmation" {
			s.l.Info("verified SNS unsubscribe confirmation", zap.String("trigger_id", trigger.ID), zap.String("topic_arn", msg.TopicARN), zap.String("message_id", msg.MessageID))
		}
	case app.TriggerAuthTypeAPIKey:
		value := ctx.GetHeader(trigger.AuthConfig.Header)
		if !strings.HasPrefix(value, trigger.AuthConfig.Prefix) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}
		value = strings.TrimPrefix(value, trigger.AuthConfig.Prefix)
		matched = matchTriggerSecret(trigger.Secrets, value, now)
		if matched == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}
	case app.TriggerAuthTypeBasic:
		username, password, ok := ctx.Request.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(trigger.AuthConfig.Username)) != 1 {
			ctx.Header("WWW-Authenticate", `Basic realm="event-ingress"`)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid basic authentication"})
			return
		}
		matched = matchTriggerSecret(trigger.Secrets, password, now)
		if matched == nil {
			ctx.Header("WWW-Authenticate", `Basic realm="event-ingress"`)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid basic authentication"})
			return
		}
	case app.TriggerAuthTypeBearerJWT:
		if err := s.verifyBearerJWT(ctx, &trigger); err != nil {
			ctx.Header("WWW-Authenticate", "Bearer")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}
	default:
		ctx.JSON(http.StatusNotImplemented, gin.H{"error": errUnsupportedAuth.Error()})
		return
	}
	if !nativeProtocolPreset {
		event, err = decodeEvent(&trigger, ctx.Request.Header, body)
		if err != nil {
			if persistErr := s.persistRejectedEvent(ctx, &trigger, body, now, "envelope decoding failed: "+err.Error()); persistErr != nil {
				ctx.Error(persistErr)
				return
			}
			ctx.JSON(http.StatusAccepted, ingressResponse{})
			return
		}
	}
	if nativeProtocolPreset {
		event, err = decodeEvent(&trigger, ctx.Request.Header, body)
		if err != nil {
			status := http.StatusBadRequest
			if trigger.Preset == "slack-events" {
				status = http.StatusOK
			}
			s.rejectIngress(ctx, &trigger, body, now, status, "envelope decoding failed: "+err.Error())
			return
		}
		if trigger.Preset == "azure-event-grid" {
			validationCode, err := azureEventGridValidationCode(event)
			if err != nil {
				s.rejectIngress(ctx, &trigger, body, now, http.StatusBadRequest, err.Error())
				return
			}
			if validationCode != "" {
				ctx.JSON(http.StatusOK, gin.H{"validationResponse": validationCode})
				return
			}
		}
		if trigger.Preset == "slack-events" {
			challenge, err := slackChallenge(event)
			if err != nil {
				s.rejectIngress(ctx, &trigger, body, now, http.StatusBadRequest, err.Error())
				return
			}
			if challenge != "" {
				ctx.JSON(http.StatusOK, gin.H{"challenge": challenge})
				return
			}
		}
	}
	if event == nil {
		ctx.JSON(http.StatusAccepted, ingressResponse{})
		return
	}
	cctx.SetOrgIDGinContext(ctx, trigger.OrgID)
	queue, err := s.orgTriggerQueue(ctx, trigger.OrgID)
	if err != nil {
		ctx.Error(err)
		return
	}
	receipt, duplicate, collision, err := s.persistEvent(ctx, &trigger, matched, queue.ID, event, body, now)
	if err != nil {
		if errors.Is(err, errTriggerInactive) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "trigger not found"})
			return
		}
		if errors.Is(err, errEventSecretInactive) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
		ctx.Error(err)
		return
	}
	if collision {
		ctx.JSON(http.StatusConflict, gin.H{"error": "event ID already exists with different content"})
		return
	}
	ctx.JSON(http.StatusAccepted, ingressResponse{EventID: receipt.ID, Duplicate: duplicate})
}

func (s *service) loadActiveTriggerSecrets(ctx context.Context, trigger *app.Trigger, now time.Time) error {
	return s.db.WithContext(ctx).
		Where(app.TriggerSecret{TriggerID: trigger.ID}).
		Where(clause.Eq{Column: "revoked_at", Value: nil}).
		Where(clause.Lte{Column: "not_before", Value: now}).
		Where(clause.Or(clause.Eq{Column: "expires_at", Value: nil}, clause.Gt{Column: "expires_at", Value: now})).
		Where(clause.Neq{Column: "secret", Value: ""}).
		Order("created_at DESC").
		Limit(maxActiveEventSecrets).
		Find(&trigger.Secrets).Error
}

func (s *service) rejectIngress(ctx *gin.Context, trigger *app.Trigger, body []byte, receivedAt time.Time, status int, reason string) {
	if err := s.persistRejectedEvent(ctx, trigger, body, receivedAt, reason); err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(status, gin.H{"error": reason})
}

func activeTriggerSecret(secret *app.TriggerSecret, now time.Time) bool {
	return secret.RevokedAt == nil && !now.Before(secret.NotBefore) && (secret.ExpiresAt == nil || now.Before(*secret.ExpiresAt))
}

func matchTriggerSecret(secrets []app.TriggerSecret, value string, now time.Time) *app.TriggerSecret {
	for i := range secrets {
		secret := &secrets[i]
		if activeTriggerSecret(secret, now) && subtle.ConstantTimeCompare([]byte(value), []byte(secret.Secret)) == 1 {
			return secret
		}
	}
	return nil
}

func (s *service) confirmSNSSubscription(ctx context.Context, msg *snsMessage) error {
	if err := validateSNSSubscribeURL(msg.SubscribeURL, msg.TopicARN, msg.Token); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, msg.SubscribeURL, nil)
	if err != nil {
		return err
	}
	client := *s.httpClient
	client.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("confirm SNS subscription: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("confirm SNS subscription: HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *service) appTriggerQueue(ctx context.Context, appID string) (*app.Queue, error) {
	ownerType := plugins.TableName(s.db, app.App{})
	var queue app.Queue
	err := s.db.WithContext(ctx).Where(app.Queue{OwnerID: appID, OwnerType: ownerType, Name: queuepkg.AppTriggersQueueName}).First(&queue).Error
	if err == nil {
		return &queue, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.appsHelpers.EnsureAppTriggerQueue(ctx, appID)
}

func (s *service) orgTriggerQueue(ctx context.Context, orgID string) (*app.Queue, error) {
	ownerType := plugins.TableName(s.db, app.Org{})
	var queue app.Queue
	err := s.db.WithContext(ctx).Where(app.Queue{OwnerID: orgID, OwnerType: ownerType, Name: queuepkg.OrgSignalsQueueName}).First(&queue).Error
	if err == nil {
		return &queue, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OrgID: &orgID, OwnerID: orgID, OwnerType: ownerType, Namespace: "orgs", Name: queuepkg.OrgSignalsQueueName,
		MaxInFlight: 10, MaxDepth: 50,
	})
}

func (s *service) persistEvent(ctx *gin.Context, trigger *app.Trigger, secret *app.TriggerSecret, queueID string, envelope *normalizedEvent, body []byte, now time.Time) (*app.TriggerEvent, bool, bool, error) {
	hash := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hash[:])
	payloadValue, err := decodeTriggerEventJSON(envelope.Payload)
	if err != nil {
		return nil, false, false, err
	}
	canonicalPayload, err := json.Marshal(payloadValue)
	if err != nil {
		return nil, false, false, err
	}
	payloadHash := sha256.Sum256(canonicalPayload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])
	normalizedType := strings.TrimSpace(envelope.Type)
	dedupeID := envelope.DedupeID
	if dedupeID == "" {
		dedupeID = envelope.ID
	}
	routingGenerationToken := uuid.NewString()
	event := app.TriggerEvent{
		TriggerID: trigger.ID, OrgID: trigger.OrgID,
		ExternalID: envelope.ID, DedupeID: dedupeID, Source: envelope.Source, EventType: normalizedType, OccurredAt: envelope.OccurredAt,
		ReceivedAt: now, Payload: envelope.Payload, PayloadContentType: envelope.ContentType,
		Headers: eventHeaders(trigger, ctx.Request.Header), RawBody: body, RawBodySHA256: hashHex, PayloadSHA256: payloadHashHex, RawBodySize: int64(len(body)), RawContentType: ctx.GetHeader("Content-Type"),
		RoutingGenerationToken: &routingGenerationToken,
	}
	if secret != nil {
		event.TriggerSecretID = &secret.ID
		event.SecretKeyID = secret.KeyID
	}
	duplicate := false
	collision := false
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedTrigger app.Trigger
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Trigger{ID: trigger.ID}).First(&lockedTrigger).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errTriggerInactive
			}
			return err
		}
		if lockedTrigger.Status != app.TriggerStatusActive {
			return errTriggerInactive
		}
		if lockedTrigger.IngressKeyHash != trigger.IngressKeyHash {
			return errTriggerInactive
		}
		var lockedSecret *app.TriggerSecret
		if secret != nil {
			lockedSecret = &app.TriggerSecret{}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.TriggerSecret{ID: secret.ID, TriggerID: trigger.ID}).First(lockedSecret).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errEventSecretInactive
				}
				return err
			}
			secretCheckTime := time.Now()
			if lockedSecret.RevokedAt != nil || secretCheckTime.Before(lockedSecret.NotBefore) || (lockedSecret.ExpiresAt != nil && !secretCheckTime.Before(*lockedSecret.ExpiresAt)) {
				return errEventSecretInactive
			}
		}
		res := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "trigger_id"}, {Name: "source"}, {Name: "dedupe_id"}}, DoNothing: true}).Create(&event)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			duplicate = true
			var existing app.TriggerEvent
			if err := tx.Unscoped().Where(app.TriggerEvent{TriggerID: trigger.ID, Source: envelope.Source, DedupeID: dedupeID}).First(&existing).Error; err != nil {
				return err
			}
			if existing.PayloadSHA256 == "" {
				existingPayload, err := decodeTriggerEventJSON(existing.Payload)
				if err != nil {
					return err
				}
				canonicalExisting, err := json.Marshal(existingPayload)
				if err != nil {
					return err
				}
				existingHash := sha256.Sum256(canonicalExisting)
				existing.PayloadSHA256 = hex.EncodeToString(existingHash[:])
				res := tx.Table(plugins.TableName(tx, &app.TriggerEvent{})).
					Where(map[string]any{"id": existing.ID}).
					Where(clause.Or(clause.Eq{Column: "payload_sha256", Value: nil}, clause.Eq{Column: "payload_sha256", Value: ""})).
					Update("payload_sha256", existing.PayloadSHA256)
				if res.Error != nil {
					return res.Error
				}
				if res.RowsAffected == 0 {
					return errors.New("payload hash changed while processing duplicate event")
				}
			}
			event = existing
			collision = event.PayloadSHA256 != payloadHashHex || event.EventType != normalizedType
		}
		if !collision && !duplicate {
			dedupeKey := "trigger-event:" + event.ID
			_, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
				QueueID: queueID, Signal: signal.NewRaw(signal.SignalType("trigger-event"), map[string]any{"event_id": event.ID, "routing_generation_token": routingGenerationToken}),
				OwnerID: event.ID, OwnerType: plugins.TableName(s.db, app.TriggerEvent{}), DedupeKey: &dedupeKey,
			})
			if err != nil {
				return err
			}
		}
		if collision {
			return nil
		}
		if lockedSecret != nil {
			if err := tx.Model(lockedSecret).Update("last_used_at", now).Error; err != nil {
				return err
			}
		}
		return tx.Model(&lockedTrigger).Update("last_event_at", now).Error
	})
	if err != nil {
		return nil, false, false, err
	}
	return &event, duplicate, collision, nil
}

func eventHeaders(trigger *app.Trigger, headers http.Header) http.Header {
	redacted := headers.Clone()
	redacted.Del("Authorization")
	redacted.Del("Proxy-Authorization")
	redacted.Del("Cookie")
	redacted.Del("Set-Cookie")
	if trigger.AuthType == app.TriggerAuthTypeHMAC || trigger.AuthType == app.TriggerAuthTypeAPIKey {
		header := trigger.AuthConfig.Header
		if trigger.AuthType == app.TriggerAuthTypeHMAC && header == "" {
			header = "X-Nuon-Signature"
		}
		redacted.Del(header)
	}
	return redacted
}

func (s *service) persistRejectedEvent(ctx *gin.Context, trigger *app.Trigger, body []byte, receivedAt time.Time, reason string) error {
	return s.persistRejectedEventWithSize(ctx, trigger, body, int64(len(body)), receivedAt, reason)
}

func (s *service) persistRejectedEventWithSize(ctx *gin.Context, trigger *app.Trigger, body []byte, rawBodySize int64, receivedAt time.Time, reason string) error {
	if body == nil {
		body = []byte{}
	}
	hash := sha256.Sum256(body)
	payloadHash := sha256.Sum256([]byte("{}"))
	completedAt := receivedAt
	event := app.TriggerEvent{
		TriggerID:          trigger.ID,
		OrgID:              trigger.OrgID,
		ExternalID:         uuid.NewString(),
		ReceivedAt:         receivedAt,
		Payload:            json.RawMessage("{}"),
		Headers:            eventHeaders(trigger, ctx.Request.Header),
		RawBody:            body,
		RawBodySHA256:      hex.EncodeToString(hash[:]),
		PayloadSHA256:      hex.EncodeToString(payloadHash[:]),
		RawBodySize:        rawBodySize,
		RawContentType:     ctx.GetHeader("Content-Type"),
		RoutingStatus:      app.EventRoutingStatusRejected,
		RoutingError:       reason,
		RoutingCompletedAt: &completedAt,
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedTrigger app.Trigger
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Trigger{ID: trigger.ID}).First(&lockedTrigger).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errTriggerInactive
			}
			return err
		}
		if lockedTrigger.Status != app.TriggerStatusActive || lockedTrigger.IngressKeyHash != trigger.IngressKeyHash {
			return errTriggerInactive
		}
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return tx.Model(&lockedTrigger).Update("last_event_at", receivedAt).Error
	})
}
