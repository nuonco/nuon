package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type createEventSourceRequest struct {
	Name        string                    `json:"name" binding:"required"`
	Description string                    `json:"description"`
	Preset      string                    `json:"preset"`
	AuthType    app.EventSourceAuthType   `json:"auth_type"`
	AuthConfig  app.EventSourceAuthConfig `json:"auth_config"`
	Envelope    app.EventEnvelopeType     `json:"envelope"`
	TypeFrom    app.EventFieldSelector    `json:"type_from"`
	IDFrom      app.EventFieldSelector    `json:"id_from"`
}

type credentialResponse struct {
	EventSource app.EventSource `json:"event_source"`
	IngressURL  string          `json:"ingress_url,omitempty"`
	KeyID       string          `json:"key_id,omitempty"`
	Secret      string          `json:"secret,omitempty"`
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

func (s *service) scopedSource(ctx context.Context, orgID, sourceID string) (*app.EventSource, error) {
	var source app.EventSource
	err := s.db.WithContext(ctx).Preload("Secrets").Where(app.EventSource{ID: sourceID, OrgID: orgID}).First(&source).Error
	return &source, err
}

// @ID CreateEventSource
// @Summary Create an event source
// @Tags event-automations
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param request body createEventSourceRequest true "Event source"
// @Success 201 {object} credentialResponse
// @Router /v1/event-sources [post]
func (s *service) CreateEventSource(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
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
	if err := applyEventSourcePreset(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.AuthType == "" {
		req.AuthType = app.EventSourceAuthTypeHMAC
	}
	if req.Envelope == "" {
		req.Envelope = app.EventEnvelopeTypeNone
	}
	if err := defaultAndValidateAuthConfig(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Envelope != app.EventEnvelopeTypeNone && req.Envelope != app.EventEnvelopeTypeCloudEvents && req.Envelope != app.EventEnvelopeTypePubSubPush && req.Envelope != app.EventEnvelopeTypeSNS {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "envelope is not implemented"})
		return
	}
	if err := validateEventFieldSelector(req.TypeFrom); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid type_from: " + err.Error()})
		return
	}
	if err := validateEventFieldSelector(req.IDFrom); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id_from: " + err.Error()})
		return
	}
	if req.Envelope == app.EventEnvelopeTypeNone {
		if req.TypeFrom == (app.EventFieldSelector{}) {
			req.TypeFrom.Header = "X-Nuon-Event-Type"
		}
		if req.IDFrom == (app.EventFieldSelector{}) {
			req.IDFrom.Header = "X-Nuon-Event-ID"
		}
	}
	ingressKey, err := generateCredential()
	if err != nil {
		ctx.Error(err)
		return
	}
	source := app.EventSource{
		OrgID: org.ID, Name: req.Name, Description: req.Description,
		IngressKeyHash: hashIngressKey(ingressKey), AuthType: req.AuthType, AuthConfig: req.AuthConfig, Envelope: req.Envelope,
		TypeFrom: req.TypeFrom, IDFrom: req.IDFrom,
	}
	var secretValue string
	var secret app.EventSourceSecret
	if authUsesSecret(req.AuthType) {
		secretValue, err = generateCredential()
		if err != nil {
			ctx.Error(err)
			return
		}
		secret = app.EventSourceSecret{OrgID: org.ID, Secret: secretValue, NotBefore: time.Now()}
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&source).Error; err != nil {
			return err
		}
		if authUsesSecret(req.AuthType) {
			secret.EventSourceID = source.ID
			return tx.Create(&secret).Error
		}
		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ingressURL, _ := url.JoinPath(s.cfg.PublicAPIURL, "v1", "event-ingress", ingressKey)
	ctx.JSON(http.StatusCreated, credentialResponse{EventSource: source, IngressURL: ingressURL, KeyID: secret.KeyID, Secret: secretValue})
}

func authUsesSecret(authType app.EventSourceAuthType) bool {
	return authType == app.EventSourceAuthTypeHMAC || authType == app.EventSourceAuthTypeAPIKey || authType == app.EventSourceAuthTypeBasic
}

func defaultAndValidateAuthConfig(req *createEventSourceRequest) error {
	switch req.AuthType {
	case app.EventSourceAuthTypeNone:
		return nil
	case app.EventSourceAuthTypeSNSSignature:
		if req.AuthConfig.TopicARN == "" {
			return errors.New("sns_signature auth requires topic_arn")
		}
		return nil
	case app.EventSourceAuthTypeHMAC:
		if req.AuthConfig.Header == "" {
			req.AuthConfig.Header = "X-Nuon-Signature"
		}
		if req.AuthConfig.Algorithm == "" {
			req.AuthConfig.Algorithm = "sha256"
		}
		if req.AuthConfig.Encoding == "" {
			req.AuthConfig.Encoding = "hex"
		}
		if req.AuthConfig.Header == "X-Nuon-Signature" && req.AuthConfig.Prefix == "" {
			req.AuthConfig.Prefix = "v1="
		}
		if req.AuthConfig.Algorithm != "sha256" && req.AuthConfig.Algorithm != "sha512" {
			return errors.New("HMAC algorithm must be sha256 or sha512")
		}
		if req.AuthConfig.Encoding != "hex" && req.AuthConfig.Encoding != "base64" {
			return errors.New("HMAC encoding must be hex or base64")
		}
		if req.AuthConfig.Header == "X-Nuon-Signature" && (req.AuthConfig.Prefix != "v1=" || req.AuthConfig.Algorithm != "sha256" || req.AuthConfig.Encoding != "hex") {
			return errors.New("X-Nuon-Signature requires prefix v1=, sha256, and hex encoding")
		}
		return nil
	case app.EventSourceAuthTypeAPIKey:
		if req.AuthConfig.Header == "" {
			req.AuthConfig.Header = "X-Nuon-API-Key"
		}
		return nil
	case app.EventSourceAuthTypeBasic:
		if req.AuthConfig.Username == "" {
			req.AuthConfig.Username = "nuon"
		}
		return nil
	case app.EventSourceAuthTypeBearerJWT:
		if req.AuthConfig.Issuer == "" || len(req.AuthConfig.Audience) == 0 {
			return errors.New("bearer_jwt auth requires issuer and audience")
		}
		if err := validateOIDCIssuer(req.AuthConfig.Issuer); err != nil {
			return err
		}
		for _, audience := range req.AuthConfig.Audience {
			if audience == "" {
				return errors.New("bearer_jwt audience may not be empty")
			}
		}
		return nil
	default:
		return errors.New("unsupported auth_type")
	}
}

func applyEventSourcePreset(req *createEventSourceRequest) error {
	if req.Preset == "" {
		return nil
	}
	type preset struct {
		auth       app.EventSourceAuthType
		envelope   app.EventEnvelopeType
		authConfig app.EventSourceAuthConfig
		typeHeader string
		idHeader   string
	}
	presets := map[string]preset{
		"github":          {auth: app.EventSourceAuthTypeHMAC, authConfig: app.EventSourceAuthConfig{Header: "X-Hub-Signature-256", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"}, typeHeader: "X-GitHub-Event", idHeader: "X-GitHub-Delivery"},
		"gitlab":          {auth: app.EventSourceAuthTypeAPIKey, authConfig: app.EventSourceAuthConfig{Header: "X-Gitlab-Token"}, typeHeader: "X-Gitlab-Event", idHeader: "X-Gitlab-Event-UUID"},
		"bitbucket":       {auth: app.EventSourceAuthTypeHMAC, authConfig: app.EventSourceAuthConfig{Header: "X-Hub-Signature", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"}, typeHeader: "X-Event-Key", idHeader: "X-Request-UUID"},
		"gitea":           {auth: app.EventSourceAuthTypeHMAC, authConfig: app.EventSourceAuthConfig{Header: "X-Gitea-Signature", Algorithm: "sha256", Encoding: "hex"}, typeHeader: "X-Gitea-Event", idHeader: "X-Gitea-Delivery"},
		"forgejo":         {auth: app.EventSourceAuthTypeHMAC, authConfig: app.EventSourceAuthConfig{Header: "X-Forgejo-Signature", Algorithm: "sha256", Encoding: "hex"}, typeHeader: "X-Forgejo-Event", idHeader: "X-Forgejo-Delivery"},
		"terraform-cloud": {auth: app.EventSourceAuthTypeHMAC, authConfig: app.EventSourceAuthConfig{Header: "X-TFE-Notification-Signature", Algorithm: "sha512", Encoding: "hex"}},
		"google-pubsub":   {auth: app.EventSourceAuthTypeBearerJWT, envelope: app.EventEnvelopeTypePubSubPush, authConfig: app.EventSourceAuthConfig{Issuer: "https://accounts.google.com"}},
		"azure-devops":    {auth: app.EventSourceAuthTypeBasic},
		"aws-eventbridge": {auth: app.EventSourceAuthTypeAPIKey},
		"aws-sns":         {auth: app.EventSourceAuthTypeSNSSignature, envelope: app.EventEnvelopeTypeSNS},
	}
	p, ok := presets[req.Preset]
	if !ok {
		return fmt.Errorf("unknown event source preset %q", req.Preset)
	}
	if req.AuthType == "" {
		req.AuthType = p.auth
	}
	if req.Envelope == "" {
		req.Envelope = p.envelope
	}
	if req.AuthConfig.Header == "" {
		req.AuthConfig.Header = p.authConfig.Header
	}
	if req.AuthConfig.Prefix == "" {
		req.AuthConfig.Prefix = p.authConfig.Prefix
	}
	if req.AuthConfig.Algorithm == "" {
		req.AuthConfig.Algorithm = p.authConfig.Algorithm
	}
	if req.AuthConfig.Encoding == "" {
		req.AuthConfig.Encoding = p.authConfig.Encoding
	}
	if req.AuthConfig.Issuer == "" {
		req.AuthConfig.Issuer = p.authConfig.Issuer
	}
	if req.TypeFrom == (app.EventFieldSelector{}) && p.typeHeader != "" {
		req.TypeFrom.Header = p.typeHeader
	}
	if req.IDFrom == (app.EventFieldSelector{}) && p.idHeader != "" {
		req.IDFrom.Header = p.idHeader
	}
	return nil
}

func validateEventFieldSelector(selector app.EventFieldSelector) error {
	if selector.Header != "" && selector.Payload != "" {
		return errors.New("exactly one of header or payload may be set")
	}
	if selector.Payload != "" {
		if _, err := eventfilter.ParsePath(selector.Payload, false); err != nil {
			return fmt.Errorf("invalid payload selector: %w", err)
		}
	}
	return nil
}

// @ID ListEventSources
// @Summary List event sources
// @Tags event-automations
// @Produce json
// @Security APIKey
// @Security OrgID
// @Success 200 {array} app.EventSource
// @Router /v1/event-sources [get]
func (s *service) ListEventSources(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var sources []app.EventSource
	if err := s.db.WithContext(ctx).Preload("Secrets").Where(app.EventSource{OrgID: org.ID}).Find(&sources).Error; err != nil {
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
// @Param event_source_id path string true "Event source ID"
// @Success 200 {object} app.EventSource
// @Router /v1/event-sources/{event_source_id} [get]
func (s *service) GetEventSource(ctx *gin.Context) {
	source, ok := s.managementSource(ctx)
	if ok {
		ctx.JSON(http.StatusOK, eventSourceAPIResponse(source))
	}
}

// @ID DeleteEventSource
// @Summary Delete an event source
// @Tags event-automations
// @Security APIKey
// @Security OrgID
// @Param event_source_id path string true "Event source ID"
// @Param force query bool false "Delete a source referenced by rules"
// @Success 204
// @Router /v1/event-sources/{event_source_id} [delete]
func (s *service) DeleteEventSource(ctx *gin.Context) {
	source, ok := s.managementSource(ctx)
	if !ok {
		return
	}
	var references int64
	if err := s.db.WithContext(ctx).Model(&app.EventAutomationRule{}).
		Where(app.EventAutomationRule{OrgID: source.OrgID, EventSourceID: source.ID}).
		Count(&references).Error; err != nil {
		ctx.Error(err)
		return
	}
	if references != 0 && ctx.Query("force") != "true" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "event source is referenced by automation rules; use force=true to delete it"})
		return
	}
	if err := s.db.WithContext(ctx).Delete(source).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *service) managementSource(ctx *gin.Context) (*app.EventSource, bool) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	source, err := s.scopedSource(ctx, org.ID, ctx.Param("event_source_id"))
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
// @Param event_source_id path string true "Event source ID"
// @Success 201 {object} credentialResponse
// @Router /v1/event-sources/{event_source_id}/rotate-secret [post]
func (s *service) RotateSecret(ctx *gin.Context) {
	source, ok := s.managementSource(ctx)
	if !ok {
		return
	}
	if !authUsesSecret(source.AuthType) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "event source does not use a managed secret"})
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
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Secrets").Where(app.EventSource{ID: source.ID, OrgID: source.OrgID}).First(&lockedSource).Error; err != nil {
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
// @Router /v1/event-sources/{event_source_id}/enable [post]
func (s *service) EnableEventSource(ctx *gin.Context) {
	s.setSourceStatus(ctx, app.EventSourceStatusActive)
}

// @ID DisableEventSource
// @Summary Disable an event source
// @Tags event-automations
// @Security APIKey
// @Security OrgID
// @Router /v1/event-sources/{event_source_id}/disable [post]
func (s *service) DisableEventSource(ctx *gin.Context) {
	s.setSourceStatus(ctx, app.EventSourceStatusSuspended)
}

// @ID RevokeEventSourceSecret
// @Summary Revoke an event source secret
// @Tags event-automations
// @Security APIKey
// @Security OrgID
// @Router /v1/event-sources/{event_source_id}/secrets/{secret_id}/revoke [post]
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
