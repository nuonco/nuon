package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type createEventSourceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type credentialResponse struct {
	EventSource app.EventSource `json:"event_source"`
	IngressURL  string          `json:"ingress_url,omitempty"`
	KeyID       string          `json:"key_id"`
	Secret      string          `json:"secret"`
}

type eventSourceSecretResponse struct {
	ID         string     `json:"id"`
	KeyID      string     `json:"key_id"`
	CreatedAt  time.Time  `json:"created_at"`
	NotBefore  time.Time  `json:"not_before"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type eventSourceResponse struct {
	app.EventSource
	Secrets []eventSourceSecretResponse `json:"secrets"`
}

func eventSourceAPIResponse(source *app.EventSource) eventSourceResponse {
	resp := eventSourceResponse{EventSource: *source, Secrets: make([]eventSourceSecretResponse, 0, len(source.Secrets))}
	for _, secret := range source.Secrets {
		resp.Secrets = append(resp.Secrets, eventSourceSecretResponse{
			ID: secret.ID, KeyID: secret.KeyID, CreatedAt: secret.CreatedAt, NotBefore: secret.NotBefore,
			ExpiresAt: secret.ExpiresAt, RevokedAt: secret.RevokedAt, LastUsedAt: secret.LastUsedAt,
		})
	}
	return resp
}

func (s *service) scopedApp(ctx context.Context, orgID, appID string) error {
	return s.db.WithContext(ctx).Where(app.App{ID: appID, OrgID: orgID}).First(&app.App{}).Error
}

func (s *service) scopedSource(ctx context.Context, orgID, appID, sourceID string) (*app.EventSource, error) {
	var source app.EventSource
	err := s.db.WithContext(ctx).Preload("Secrets").Where(app.EventSource{ID: sourceID, OrgID: orgID, AppID: appID}).First(&source).Error
	return &source, err
}

// @ID CreateEventSource
// @Summary Create an event source
// @Tags event-automations
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "App ID"
// @Param request body createEventSourceRequest true "Event source"
// @Success 201 {object} credentialResponse
// @Router /v1/apps/{app_id}/event-sources [post]
func (s *service) CreateEventSource(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	appID := ctx.Param("app_id")
	if err := s.scopedApp(ctx, org.ID, appID); err != nil {
		ctx.Error(fmt.Errorf("app not found: %w", err))
		return
	}
	var req createEventSourceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 128 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name must be between 1 and 128 characters"})
		return
	}
	ingressKey, err := generateCredential()
	if err != nil {
		ctx.Error(err)
		return
	}
	secretValue, err := generateCredential()
	if err != nil {
		ctx.Error(err)
		return
	}
	if _, err := s.appsHelpers.EnsureAppAutomationQueue(ctx, appID); err != nil {
		ctx.Error(err)
		return
	}
	source := app.EventSource{OrgID: org.ID, AppID: appID, Name: req.Name, Description: req.Description, IngressKeyHash: hashIngressKey(ingressKey)}
	secret := app.EventSourceSecret{OrgID: org.ID, Secret: secretValue, NotBefore: time.Now()}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&source).Error; err != nil {
			return err
		}
		secret.EventSourceID = source.ID
		return tx.Create(&secret).Error
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ingressURL, _ := url.JoinPath(s.cfg.PublicAPIURL, "v1", "event-ingress", ingressKey)
	ctx.JSON(http.StatusCreated, credentialResponse{EventSource: source, IngressURL: ingressURL, KeyID: secret.KeyID, Secret: secretValue})
}

// @ID ListEventSources
// @Summary List event sources
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "App ID"
// @Success 200 {array} app.EventSource
// @Router /v1/apps/{app_id}/event-sources [get]
func (s *service) ListEventSources(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	appID := ctx.Param("app_id")
	if err := s.scopedApp(ctx, org.ID, appID); err != nil {
		ctx.Error(err)
		return
	}
	var sources []app.EventSource
	if err := s.db.WithContext(ctx).Preload("Secrets").Where(app.EventSource{OrgID: org.ID, AppID: appID}).Find(&sources).Error; err != nil {
		ctx.Error(err)
		return
	}
	resp := make([]eventSourceResponse, 0, len(sources))
	for i := range sources {
		resp = append(resp, eventSourceAPIResponse(&sources[i]))
	}
	ctx.JSON(http.StatusOK, resp)
}

// @ID GetEventSource
// @Summary Get an event source
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "App ID"
// @Param event_source_id path string true "Event source ID"
// @Success 200 {object} app.EventSource
// @Router /v1/apps/{app_id}/event-sources/{event_source_id} [get]
func (s *service) GetEventSource(ctx *gin.Context) {
	source, ok := s.managementSource(ctx)
	if ok {
		ctx.JSON(http.StatusOK, eventSourceAPIResponse(source))
	}
}

func (s *service) managementSource(ctx *gin.Context) (*app.EventSource, bool) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	appID := ctx.Param("app_id")
	if err := s.scopedApp(ctx, org.ID, appID); err != nil {
		ctx.Error(err)
		return nil, false
	}
	source, err := s.scopedSource(ctx, org.ID, appID, ctx.Param("event_source_id"))
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	return source, true
}

// @ID RotateEventSourceSecret
// @Summary Rotate an event source secret
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param app_id path string true "App ID"
// @Param event_source_id path string true "Event source ID"
// @Success 201 {object} credentialResponse
// @Router /v1/apps/{app_id}/event-sources/{event_source_id}/rotate-secret [post]
func (s *service) RotateSecret(ctx *gin.Context) {
	source, ok := s.managementSource(ctx)
	if !ok {
		return
	}
	value, err := generateCredential()
	if err != nil {
		ctx.Error(err)
		return
	}
	now := time.Now()
	grace := now.Add(24 * time.Hour)
	secret := app.EventSourceSecret{OrgID: source.OrgID, EventSourceID: source.ID, Secret: value, NotBefore: now}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedSource app.EventSource
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Secrets").Where(app.EventSource{ID: source.ID, OrgID: source.OrgID, AppID: source.AppID}).First(&lockedSource).Error; err != nil {
			return err
		}
		for i := range lockedSource.Secrets {
			previous := &lockedSource.Secrets[i]
			if previous.RevokedAt == nil && (previous.ExpiresAt == nil || previous.ExpiresAt.After(grace)) {
				if err := tx.Model(previous).Update("expires_at", grace).Error; err != nil {
					return err
				}
			}
		}
		return tx.Create(&secret).Error
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, credentialResponse{EventSource: *source, KeyID: secret.KeyID, Secret: value})
}

func (s *service) setSourceStatus(ctx *gin.Context, status app.EventSourceStatus) {
	source, ok := s.managementSource(ctx)
	if !ok {
		return
	}
	if err := s.db.WithContext(ctx).Model(source).Update("status", status).Error; err != nil {
		ctx.Error(err)
		return
	}
	source.Status = status
	ctx.JSON(http.StatusOK, source)
}

// @ID EnableEventSource
// @Summary Enable an event source
// @Tags event-automations
// @Security APIKey
// @Security OrgID
// @Router /v1/apps/{app_id}/event-sources/{event_source_id}/enable [post]
func (s *service) EnableEventSource(ctx *gin.Context) {
	s.setSourceStatus(ctx, app.EventSourceStatusActive)
}

// @ID DisableEventSource
// @Summary Disable an event source
// @Tags event-automations
// @Security APIKey
// @Security OrgID
// @Router /v1/apps/{app_id}/event-sources/{event_source_id}/disable [post]
func (s *service) DisableEventSource(ctx *gin.Context) {
	s.setSourceStatus(ctx, app.EventSourceStatusSuspended)
}

// @ID RevokeEventSourceSecret
// @Summary Revoke an event source secret
// @Tags event-automations
// @Security APIKey
// @Security OrgID
// @Router /v1/apps/{app_id}/event-sources/{event_source_id}/secrets/{secret_id}/revoke [post]
func (s *service) RevokeSecret(ctx *gin.Context) {
	source, ok := s.managementSource(ctx)
	if !ok {
		return
	}
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&app.EventSourceSecret{}).Where(app.EventSourceSecret{ID: ctx.Param("secret_id"), OrgID: source.OrgID, EventSourceID: source.ID}).Update("revoked_at", now)
	if res.Error != nil {
		ctx.Error(res.Error)
		return
	}
	if res.RowsAffected == 0 {
		ctx.Error(gorm.ErrRecordNotFound)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"revoked_at": now})
}
