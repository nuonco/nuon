package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
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
// @Summary Ingest a CloudEvent
// @Tags event-automations
// @Accept application/cloudevents+json
// @Produce json
// @Param ingress_key path string true "Opaque ingress key"
// @Success 202 {object} ingressResponse
// @Router /v1/event-ingress/{ingress_key} [post]
func (s *service) IngestEvent(ctx *gin.Context) {
	mediaType, _, err := mime.ParseMediaType(ctx.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/cloudevents+json" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "content type must be application/cloudevents+json"})
		return
	}
	timestamp, err := parseTimestamp(ctx.GetHeader("X-Nuon-Timestamp"), time.Now())
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
	var matched *app.EventSourceSecret
	for i := range source.Secrets {
		secret := &source.Secrets[i]
		if secret.RevokedAt != nil || now.Before(secret.NotBefore) || (secret.ExpiresAt != nil && !now.Before(*secret.ExpiresAt)) {
			continue
		}
		if verifySignature(secret.Secret, timestamp, body, signature) {
			matched = secret
			break
		}
	}
	if matched == nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}
	event, err := parseCloudEvent(body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid CloudEvent"})
		return
	}
	cctx.SetOrgIDGinContext(ctx, source.OrgID)
	queue, err := s.appAutomationQueue(ctx, source.AppID)
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

func (s *service) persistEvent(ctx *gin.Context, source *app.EventSource, secret *app.EventSourceSecret, queueID string, envelope *cloudEvent, body []byte, now time.Time) (*app.EventSourceEvent, bool, bool, error) {
	hash := sha256.Sum256(body)
	hashHex := hex.EncodeToString(hash[:])
	event := app.EventSourceEvent{
		EventSourceID: source.ID, EventSourceSecretID: secret.ID, AppID: source.AppID, OrgID: source.OrgID,
		ExternalID: envelope.ID, CloudEventSource: envelope.Source, EventType: envelope.Type, Subject: envelope.Subject, OccurredAt: envelope.Time,
		ReceivedAt: now, Payload: envelope.Data, PayloadSHA256: hashHex, PayloadContentType: envelope.DataContentType,
		PayloadSize: int64(len(body)), SecretKeyID: secret.KeyID,
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
		var lockedSecret app.EventSourceSecret
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.EventSourceSecret{ID: secret.ID, EventSourceID: source.ID}).First(&lockedSecret).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errEventSecretInactive
			}
			return err
		}
		secretCheckTime := time.Now()
		if lockedSecret.RevokedAt != nil || secretCheckTime.Before(lockedSecret.NotBefore) || (lockedSecret.ExpiresAt != nil && !secretCheckTime.Before(*lockedSecret.ExpiresAt)) {
			return errEventSecretInactive
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
			collision = event.PayloadSHA256 != hashHex
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
		if err := tx.Model(&lockedSecret).Update("last_used_at", now).Error; err != nil {
			return err
		}
		return tx.Model(&lockedSource).Update("last_event_at", now).Error
	})
	if err != nil {
		return nil, false, false, err
	}
	return &event, duplicate, collision, nil
}
