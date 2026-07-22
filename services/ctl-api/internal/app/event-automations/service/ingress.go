package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

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
		if eventID == "" && source.AuthType == app.EventSourceAuthTypeNone {
			eventID = uuid.NewString()
		}
		return &normalizedEvent{
			ID: eventID, Type: headers.Get(source.TypeFrom.Header), Payload: json.RawMessage(body), ContentType: headers.Get("Content-Type"),
		}, nil
	case app.EventEnvelopeTypeCloudEvents:
		event, err := parseCloudEvent(body)
		if err != nil {
			return nil, err
		}
		return &normalizedEvent{ID: event.ID, Type: event.Type, OccurredAt: event.Time, Payload: event.Data, ContentType: event.DataContentType}, nil
	default:
		return nil, errUnsupportedEnvelope
	}
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
		signedPayload, err := hmacPayload(source.Envelope, event.ID, event.Type, body)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		timestamp, err := parseTimestamp(ctx.GetHeader("X-Nuon-Timestamp"), now)
		if err != nil {
			if errors.Is(err, errStaleTimestamp) {
				ctx.JSON(http.StatusUnauthorized, gin.H{"error": "stale signature"})
				return
			}
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid timestamp"})
			return
		}
		signature, err := parseSignature(ctx.GetHeader("X-Nuon-Signature"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid signature header"})
			return
		}
		for i := range source.Secrets {
			secret := &source.Secrets[i]
			if secret.RevokedAt != nil || now.Before(secret.NotBefore) || (secret.ExpiresAt != nil && !now.Before(*secret.ExpiresAt)) {
				continue
			}
			if verifySignature(secret.Secret, timestamp, signedPayload, signature) {
				matched = secret
				break
			}
		}
		if matched == nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
	default:
		ctx.JSON(http.StatusNotImplemented, gin.H{"error": errUnsupportedAuth.Error()})
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
		Headers: ctx.Request.Header.Clone(), RawBody: body, RawBodySHA256: hashHex, RawBodySize: int64(len(body)), RawContentType: ctx.GetHeader("Content-Type"),
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
