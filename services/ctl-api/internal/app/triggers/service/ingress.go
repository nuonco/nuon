package service

import (
	"context"
	"crypto/sha256"
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
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/events/envelope"
	"github.com/nuonco/nuon/pkg/events/provider"
	"github.com/nuonco/nuon/pkg/events/providers"
	eventsns "github.com/nuonco/nuon/pkg/events/sns"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queuepkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const (
	maxEventBody             = 16 * 1024 * 1024
	maxActiveEventSecrets    = 2
	maxEventSecretMetadata   = 20
	eventSecretRotationGrace = 24 * time.Hour
)

var (
	errEventBodyTooLarge   = errors.New("request body too large")
	errTriggerInactive     = errors.New("trigger inactive")
	errEventSecretInactive = errors.New("trigger secret inactive")
	errUnsupportedAuth     = errors.New("trigger auth is not implemented")
)

type ingressResponse struct {
	EventID   string `json:"event_id"`
	Duplicate bool   `json:"duplicate"`
}

func decodeEvent(trigger *app.Trigger, headers http.Header, body []byte) (*envelope.Event, error) {
	decoder, err := providers.Resolve(trigger.Preset).Decoder(provider.EnvelopeType(trigger.Envelope))
	if err != nil {
		return nil, err
	}
	event, err := decoder.Decode(headers, body)
	if err != nil || event == nil {
		return nil, err
	}
	if err := envelope.ApplySelectors(event, headers, envelope.FieldSelector(trigger.TypeFrom), envelope.FieldSelector(trigger.IDFrom)); err != nil {
		return nil, err
	}
	if event.DedupeID == "" {
		event.DedupeID = event.ID
	}
	return event, nil
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
	prov := providers.Resolve(trigger.Preset)
	var matched *app.TriggerSecret
	switch trigger.AuthType {
	case app.TriggerAuthTypeNone:
	case app.TriggerAuthTypeHMAC, app.TriggerAuthTypeAPIKey, app.TriggerAuthTypeBasic:
		verifier := prov.Verifier(provider.AuthType(trigger.AuthType), provider.AuthConfig(trigger.AuthConfig))
		if verifier == nil {
			ctx.JSON(http.StatusNotImplemented, gin.H{"error": errUnsupportedAuth.Error()})
			return
		}
		active := make([]*app.TriggerSecret, 0, len(trigger.Secrets))
		values := make([]string, 0, len(trigger.Secrets))
		for i := range trigger.Secrets {
			secret := &trigger.Secrets[i]
			if activeTriggerSecret(secret, now) {
				active = append(active, secret)
				values = append(values, secret.Secret)
			}
		}
		idx, err := verifier.Verify(ctx.Request.Header, body, values, now)
		if err != nil {
			if trigger.AuthType == app.TriggerAuthTypeBasic {
				ctx.Header("WWW-Authenticate", `Basic realm="event-ingress"`)
			}
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		matched = active[idx]
	case app.TriggerAuthTypeSNSSignature:
		msg, err := eventsns.ParseMessage(body)
		if err != nil || msg.TopicARN != trigger.AuthConfig.TopicARN || s.snsVerifier.Verify(ctx, msg) != nil {
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
	event, err := decodeEvent(&trigger, ctx.Request.Header, body)
	if err != nil {
		reason := "envelope decoding failed: " + err.Error()
		if status := prov.RejectStatus(err); status != http.StatusAccepted {
			s.rejectIngress(ctx, &trigger, body, now, status, reason)
			return
		}
		if persistErr := s.persistRejectedEvent(ctx, &trigger, body, now, reason); persistErr != nil {
			ctx.Error(persistErr)
			return
		}
		ctx.JSON(http.StatusAccepted, ingressResponse{})
		return
	}
	handshake, err := prov.Handshake(event)
	if err != nil {
		s.rejectIngress(ctx, &trigger, body, now, http.StatusBadRequest, err.Error())
		return
	}
	if handshake != nil {
		ctx.JSON(handshake.Status, handshake.Body)
		return
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

func (s *service) confirmSNSSubscription(ctx context.Context, msg *eventsns.Message) error {
	if err := eventsns.ValidateSubscribeURL(msg.SubscribeURL, msg.TopicARN, msg.Token); err != nil {
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

func (s *service) persistEvent(ctx *gin.Context, trigger *app.Trigger, secret *app.TriggerSecret, queueID string, normalized *envelope.Event, body []byte, now time.Time) (*app.TriggerEvent, bool, bool, error) {
	hash := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hash[:])
	payloadValue, err := envelope.DecodeJSON(normalized.Payload)
	if err != nil {
		return nil, false, false, err
	}
	canonicalPayload, err := json.Marshal(payloadValue)
	if err != nil {
		return nil, false, false, err
	}
	payloadHash := sha256.Sum256(canonicalPayload)
	payloadHashHex := hex.EncodeToString(payloadHash[:])
	normalizedType := strings.TrimSpace(normalized.Type)
	dedupeID := normalized.DedupeID
	if dedupeID == "" {
		dedupeID = normalized.ID
	}
	routingGenerationToken := uuid.NewString()
	event := app.TriggerEvent{
		TriggerID: trigger.ID, OrgID: trigger.OrgID,
		ExternalID: normalized.ID, DedupeID: dedupeID, Source: normalized.Source, EventType: normalizedType, OccurredAt: normalized.OccurredAt,
		ReceivedAt: now, Payload: normalized.Payload, PayloadContentType: normalized.ContentType,
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
			if err := tx.Unscoped().Where(app.TriggerEvent{TriggerID: trigger.ID, Source: normalized.Source, DedupeID: dedupeID}).First(&existing).Error; err != nil {
				return err
			}
			if existing.PayloadSHA256 == "" {
				existingPayload, err := envelope.DecodeJSON(existing.Payload)
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
