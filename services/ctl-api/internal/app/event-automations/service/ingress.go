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
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	orghelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/orgs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queuepkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const maxEventBody = 256 * 1024

var (
	errEventBodyTooLarge   = errors.New("request body too large")
	errEventSourceInactive = errors.New("event source inactive")
	errEventSecretInactive = errors.New("event source secret inactive")
	errUnsupportedAuth     = errors.New("event source auth is not implemented")
	errUnsupportedEnvelope = errors.New("event source envelope is not implemented")
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
	Type        string
	OccurredAt  *time.Time
	Payload     json.RawMessage
	ContentType string
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

func decodeEvent(source *app.EventSource, headers http.Header, body []byte) (*normalizedEvent, error) {
	switch source.Envelope {
	case app.EventEnvelopeTypeNone:
		if !json.Valid(body) {
			return nil, errors.New("invalid JSON event")
		}
		eventID := headers.Get(source.IDFrom.Header)
		eventType := headers.Get(source.TypeFrom.Header)
		if source.IDFrom.Payload != "" || source.TypeFrom.Payload != "" {
			payload, err := decodeAutomationJSON(body)
			if err != nil {
				return nil, err
			}
			if source.IDFrom.Payload != "" {
				eventID, err = selectEventString(payload, source.IDFrom.Payload)
				if err != nil {
					return nil, fmt.Errorf("extract event ID: %w", err)
				}
			}
			if source.TypeFrom.Payload != "" {
				eventType, err = selectEventString(payload, source.TypeFrom.Payload)
				if err != nil {
					return nil, fmt.Errorf("extract event type: %w", err)
				}
			}
		}
		if eventID == "" && (source.AuthType != app.EventSourceAuthTypeHMAC || (source.AuthConfig.Header != "" && source.AuthConfig.Header != "X-Nuon-Signature")) {
			eventID = uuid.NewString()
		}
		return &normalizedEvent{
			ID: eventID, Type: eventType, Payload: json.RawMessage(body), ContentType: headers.Get("Content-Type"),
		}, nil
	case app.EventEnvelopeTypeCloudEvents:
		event, err := parseCloudEvent(body)
		if err != nil {
			return nil, err
		}
		return &normalizedEvent{ID: event.ID, Type: event.Type, OccurredAt: event.Time, Payload: event.Data, ContentType: event.DataContentType}, nil
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
		return &normalizedEvent{ID: push.Message.MessageID, OccurredAt: occurredAt, Payload: payload, ContentType: "application/json"}, nil
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
		return &normalizedEvent{ID: msg.MessageID, OccurredAt: &occurredAt, Payload: payload, ContentType: "application/json"}, nil
	default:
		return nil, errUnsupportedEnvelope
	}
}

func decodeAutomationJSON(body []byte) (any, error) {
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

// @ID IngestAutomationEvent
// @Summary Ingest an event
// @Tags event-automations
// @Accept json
// @Produce json
// @Param ingress_key path string true "Opaque ingress key"
// @Success 202 {object} ingressResponse
// @Router /v1/event-ingress/{ingress_key} [post]
func (s *service) IngestEvent(ctx *gin.Context) {
	body, err := readLimitedBody(ctx.Request.Body)
	if err != nil {
		if errors.Is(err, errEventBodyTooLarge) {
			ctx.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		ctx.Error(err)
		return
	}

	var source app.EventSource
	err = s.db.WithContext(ctx).Preload("Secrets").Where(app.EventSource{IngressKeyHash: hashIngressKey(ctx.Param("ingress_key")), Status: app.EventSourceStatusActive}).First(&source).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "event source not found"})
			return
		}
		ctx.Error(err)
		return
	}
	now := time.Now()
	event, err := decodeEvent(&source, ctx.Request.Header, body)
	if err != nil {
		if errors.Is(err, errUnsupportedEnvelope) {
			ctx.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid event envelope"})
		return
	}
	var matched *app.EventSourceSecret
	switch source.AuthType {
	case app.EventSourceAuthTypeNone:
	case app.EventSourceAuthTypeHMAC:
		config := source.AuthConfig
		if config.Header == "" {
			config = app.EventSourceAuthConfig{Header: "X-Nuon-Signature", Prefix: "v1=", Algorithm: "sha256", Encoding: "hex"}
		}
		isNuonSignature := config.Header == "X-Nuon-Signature" && config.Prefix == "v1="
		signedPayload := body
		if isNuonSignature {
			signedPayload, err = hmacPayload(source.Envelope, event.ID, event.Type, body)
			if err != nil {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
		}
		signature, err := decodeHMACSignature(ctx.GetHeader(config.Header), config)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature header"})
			return
		}
		var timestamp string
		if isNuonSignature {
			timestamp, err = parseTimestamp(ctx.GetHeader("X-Nuon-Timestamp"), now)
			if err != nil {
				if errors.Is(err, errStaleTimestamp) {
					ctx.JSON(http.StatusUnauthorized, gin.H{"error": "stale signature"})
					return
				}
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
				return
			}
		}
		for i := range source.Secrets {
			secret := &source.Secrets[i]
			if !activeEventSourceSecret(secret, now) {
				continue
			}
			valid := verifyGenericHMAC(secret.Secret, signedPayload, signature, config.Algorithm)
			if isNuonSignature {
				valid = verifySignature(secret.Secret, timestamp, signedPayload, signature)
			}
			if valid {
				matched = secret
				break
			}
		}
		if matched == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	case app.EventSourceAuthTypeSNSSignature:
		msg, err := parseSNSMessage(body)
		if err != nil || msg.TopicARN != source.AuthConfig.TopicARN || s.snsVerifier.verify(ctx, msg) != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid SNS signature"})
			return
		}
		if msg.Type == "SubscriptionConfirmation" {
			if err := s.confirmSNSSubscription(ctx, msg); err != nil {
				ctx.Error(err)
				return
			}
		}
	case app.EventSourceAuthTypeAPIKey:
		value := ctx.GetHeader(source.AuthConfig.Header)
		if !strings.HasPrefix(value, source.AuthConfig.Prefix) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}
		value = strings.TrimPrefix(value, source.AuthConfig.Prefix)
		matched = matchEventSourceSecret(source.Secrets, value, now)
		if matched == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid API key"})
			return
		}
	case app.EventSourceAuthTypeBasic:
		username, password, ok := ctx.Request.BasicAuth()
		if !ok || subtle.ConstantTimeCompare([]byte(username), []byte(source.AuthConfig.Username)) != 1 {
			ctx.Header("WWW-Authenticate", `Basic realm="event-ingress"`)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid basic authentication"})
			return
		}
		matched = matchEventSourceSecret(source.Secrets, password, now)
		if matched == nil {
			ctx.Header("WWW-Authenticate", `Basic realm="event-ingress"`)
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid basic authentication"})
			return
		}
	case app.EventSourceAuthTypeBearerJWT:
		if err := s.verifyBearerJWT(ctx, &source); err != nil {
			ctx.Header("WWW-Authenticate", "Bearer")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid bearer token"})
			return
		}
	default:
		ctx.JSON(http.StatusNotImplemented, gin.H{"error": errUnsupportedAuth.Error()})
		return
	}
	if event == nil {
		ctx.JSON(http.StatusAccepted, ingressResponse{})
		return
	}
	cctx.SetOrgIDGinContext(ctx, source.OrgID)
	queue, err := s.orgAutomationQueue(ctx, source.OrgID)
	if err != nil {
		ctx.Error(err)
		return
	}
	receipt, duplicate, collision, err := s.persistEvent(ctx, &source, matched, queue.ID, event, body, now)
	if err != nil {
		if errors.Is(err, errEventSourceInactive) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "event source not found"})
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

func activeEventSourceSecret(secret *app.EventSourceSecret, now time.Time) bool {
	return secret.RevokedAt == nil && !now.Before(secret.NotBefore) && (secret.ExpiresAt == nil || now.Before(*secret.ExpiresAt))
}

func matchEventSourceSecret(secrets []app.EventSourceSecret, value string, now time.Time) *app.EventSourceSecret {
	for i := range secrets {
		secret := &secrets[i]
		if activeEventSourceSecret(secret, now) && subtle.ConstantTimeCompare([]byte(value), []byte(secret.Secret)) == 1 {
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

func (s *service) appAutomationQueue(ctx context.Context, appID string) (*app.Queue, error) {
	ownerType := plugins.TableName(s.db, app.App{})
	var queue app.Queue
	err := s.db.WithContext(ctx).Where(app.Queue{OwnerID: appID, OwnerType: ownerType, Name: queuepkg.AppAutomationsQueueName}).First(&queue).Error
	if err == nil {
		return &queue, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.appsHelpers.EnsureAppAutomationQueue(ctx, appID)
}

func (s *service) orgAutomationQueue(ctx context.Context, orgID string) (*app.Queue, error) {
	ownerType := plugins.TableName(s.db, app.Org{})
	var queue app.Queue
	err := s.db.WithContext(ctx).Where(app.Queue{OwnerID: orgID, OwnerType: ownerType, Name: orghelpers.OrgSignalsQueueName}).First(&queue).Error
	if err == nil {
		return &queue, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	return s.queueClient.Create(ctx, &queueclient.CreateQueueRequest{
		OrgID: &orgID, OwnerID: orgID, OwnerType: ownerType, Namespace: "orgs", Name: orghelpers.OrgSignalsQueueName,
		MaxInFlight: 10, MaxDepth: 50,
	})
}

func (s *service) persistEvent(ctx *gin.Context, source *app.EventSource, secret *app.EventSourceSecret, queueID string, envelope *normalizedEvent, body []byte, now time.Time) (*app.EventSourceEvent, bool, bool, error) {
	hash := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hash[:])
	event := app.EventSourceEvent{
		EventSourceID: source.ID, OrgID: source.OrgID,
		ExternalID: envelope.ID, EventType: envelope.Type, OccurredAt: envelope.OccurredAt,
		ReceivedAt: now, Payload: envelope.Payload, PayloadContentType: envelope.ContentType,
		Headers: eventHeaders(source, ctx.Request.Header), RawBody: body, RawBodySHA256: hashHex, RawBodySize: int64(len(body)), RawContentType: ctx.GetHeader("Content-Type"),
	}
	if secret != nil {
		event.EventSourceSecretID = &secret.ID
		event.SecretKeyID = secret.KeyID
	}
	duplicate := false
	collision := false
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedSource app.EventSource
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.EventSource{ID: source.ID}).First(&lockedSource).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errEventSourceInactive
			}
			return err
		}
		if lockedSource.Status != app.EventSourceStatusActive {
			return errEventSourceInactive
		}
		var lockedSecret *app.EventSourceSecret
		if secret != nil {
			lockedSecret = &app.EventSourceSecret{}
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.EventSourceSecret{ID: secret.ID, EventSourceID: source.ID}).First(lockedSecret).Error; err != nil {
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
		res := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_source_id"}, {Name: "external_id"}}, DoNothing: true}).Create(&event)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			duplicate = true
			if err := tx.Unscoped().Where(app.EventSourceEvent{EventSourceID: source.ID, ExternalID: envelope.ID}).First(&event).Error; err != nil {
				return err
			}
			collision = event.RawBodySHA256 != hashHex || event.EventType != envelope.Type
		}
		if !collision {
			dedupeKey := "automation-event:" + event.ID
			_, err := s.queueClient.EnqueueSignalInTransaction(ctx, tx, &queueclient.EnqueueSignalRequest{
				QueueID: queueID, Signal: signal.NewRaw(signal.SignalType("automation-event"), map[string]any{"event_id": event.ID}),
				OwnerID: event.ID, OwnerType: plugins.TableName(s.db, app.EventSourceEvent{}), DedupeKey: &dedupeKey,
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
		return tx.Model(&lockedSource).Update("last_event_at", now).Error
	})
	if err != nil {
		return nil, false, false, err
	}
	return &event, duplicate, collision, nil
}

func eventHeaders(source *app.EventSource, headers http.Header) http.Header {
	redacted := headers.Clone()
	redacted.Del("Authorization")
	redacted.Del("Proxy-Authorization")
	if source.AuthType == app.EventSourceAuthTypeHMAC || source.AuthType == app.EventSourceAuthTypeAPIKey {
		redacted.Del(source.AuthConfig.Header)
	}
	return redacted
}
