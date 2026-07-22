package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

var errTriggerReferenced = errors.New("trigger is referenced")

type createTriggerRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Preset      string                 `json:"preset"`
	Secret      string                 `json:"secret"`
	AuthType    app.TriggerAuthType    `json:"auth_type"`
	AuthConfig  app.TriggerAuthConfig  `json:"auth_config"`
	Envelope    app.EventEnvelopeType  `json:"envelope"`
	TypeFrom    app.EventFieldSelector `json:"type_from"`
	IDFrom      app.EventFieldSelector `json:"id_from"`
}

type credentialResponse struct {
	Trigger    app.Trigger `json:"trigger"`
	IngressURL string      `json:"ingress_url,omitempty"`
	KeyID      string      `json:"key_id,omitempty"`
	Secret     string      `json:"secret,omitempty"`
}

type triggerSecretResponse struct {
	ID         string     `json:"id"`
	KeyID      string     `json:"key_id"`
	CreatedAt  time.Time  `json:"created_at"`
	NotBefore  time.Time  `json:"not_before"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	RevokedAt  *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty"`
}

type revokeTriggerSecretResponse struct {
	RevokedAt time.Time `json:"revoked_at"`
}

type revealTriggerSecretResponse struct {
	KeyID  string `json:"key_id"`
	Secret string `json:"secret"`
}

type triggerResponse struct {
	app.Trigger
	Secrets []triggerSecretResponse `json:"secrets"`
}

type triggerIngressURLResponse struct {
	IngressURL string `json:"ingress_url,omitempty"`
}

func buildIngressURL(publicAPIURL, ingressKey string) string {
	if ingressKey == "" {
		return ""
	}
	ingressURL, _ := url.JoinPath(publicAPIURL, "v1", "event-ingress", ingressKey)
	return ingressURL
}

func triggerMetadataResponse(trigger *app.Trigger) triggerResponse {
	count := min(len(trigger.Secrets), maxEventSecretMetadata)
	resp := triggerResponse{Trigger: *trigger, Secrets: make([]triggerSecretResponse, 0, count)}
	for _, secret := range trigger.Secrets[:count] {
		resp.Secrets = append(resp.Secrets, triggerSecretResponse{
			ID: secret.ID, KeyID: secret.KeyID, CreatedAt: secret.CreatedAt, NotBefore: secret.NotBefore,
			ExpiresAt: secret.ExpiresAt, RevokedAt: secret.RevokedAt, LastUsedAt: secret.LastUsedAt,
		})
	}
	return resp
}

func (s *service) scopedTrigger(ctx context.Context, orgID, triggerID string) (*app.Trigger, error) {
	var trigger app.Trigger
	if err := s.db.WithContext(ctx).Where(app.Trigger{ID: triggerID, OrgID: orgID}).First(&trigger).Error; err != nil {
		return &trigger, err
	}
	err := s.db.WithContext(ctx).Where(app.TriggerSecret{TriggerID: trigger.ID}).Order("created_at DESC").Limit(maxEventSecretMetadata).Find(&trigger.Secrets).Error
	return &trigger, err
}

func (s *service) scrubInactiveTriggerSecrets(ctx context.Context, db *gorm.DB, triggerID string, now time.Time) error {
	return db.WithContext(ctx).Model(&app.TriggerSecret{}).
		Where(app.TriggerSecret{TriggerID: triggerID}).
		Where(clause.Neq{Column: "secret", Value: ""}).
		Where(clause.Or(clause.Neq{Column: "revoked_at", Value: nil}, clause.Lte{Column: "expires_at", Value: now})).
		Update("secret", "").Error
}

// @ID CreateTrigger
// @Summary Create an trigger
// @Tags triggers
// @Accept json
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param request body createTriggerRequest true "Trigger"
// @Success 201 {object} credentialResponse
// @Router /v1/triggers [post]
func (s *service) CreateTrigger(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var req createTriggerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" || len(req.Name) > 128 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "name must be between 1 and 128 characters"})
		return
	}
	if err := applyTriggerPreset(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Preset == "slack-events" {
		req.Secret = strings.TrimSpace(req.Secret)
		if req.Secret == "" || len(req.Secret) > 1024 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "slack-events preset requires a Slack signing secret"})
			return
		}
	} else if req.Secret != "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "secret may only be supplied for the slack-events preset"})
		return
	}
	if req.AuthType == "" {
		req.AuthType = app.TriggerAuthTypeHMAC
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
	nativeProtocolPreset := req.Preset == "azure-event-grid" || req.Preset == "slack-events"
	if req.Envelope == app.EventEnvelopeTypeNone && !nativeProtocolPreset {
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
	trigger := app.Trigger{
		OrgID: org.ID, Name: req.Name, Description: req.Description, Preset: req.Preset,
		IngressKey: ingressKey, IngressKeyHash: hashIngressKey(ingressKey), AuthType: req.AuthType, AuthConfig: req.AuthConfig, Envelope: req.Envelope,
		TypeFrom: req.TypeFrom, IDFrom: req.IDFrom,
	}
	var secretValue string
	var secret app.TriggerSecret
	if authUsesSecret(req.AuthType) {
		secretValue = req.Secret
		if secretValue == "" {
			secretValue, err = generateCredential()
			if err != nil {
				ctx.Error(err)
				return
			}
		}
		secret = app.TriggerSecret{OrgID: org.ID, Secret: secretValue, NotBefore: time.Now()}
	}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&trigger).Error; err != nil {
			return err
		}
		if authUsesSecret(req.AuthType) {
			secret.TriggerID = trigger.ID
			return tx.Create(&secret).Error
		}
		return nil
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	responseSecret := secretValue
	if req.Preset == "slack-events" {
		responseSecret = ""
	}
	ctx.JSON(http.StatusCreated, credentialResponse{Trigger: trigger, IngressURL: buildIngressURL(s.cfg.PublicAPIURL, ingressKey), KeyID: secret.KeyID, Secret: responseSecret})
}

func authUsesSecret(authType app.TriggerAuthType) bool {
	return authType == app.TriggerAuthTypeHMAC || authType == app.TriggerAuthTypeAPIKey || authType == app.TriggerAuthTypeBasic
}

func defaultAndValidateAuthConfig(req *createTriggerRequest) error {
	switch req.AuthType {
	case app.TriggerAuthTypeNone:
		return nil
	case app.TriggerAuthTypeSNSSignature:
		if req.AuthConfig.TopicARN == "" {
			return errors.New("sns_signature auth requires topic_arn")
		}
		return nil
	case app.TriggerAuthTypeHMAC:
		if req.AuthConfig.Header == "" {
			req.AuthConfig.Header = "X-Nuon-Signature"
			if req.AuthConfig.Prefix == "" {
				req.AuthConfig.Prefix = "v1="
			}
		}
		if req.AuthConfig.Algorithm == "" {
			req.AuthConfig.Algorithm = "sha256"
		}
		if req.AuthConfig.Encoding == "" {
			req.AuthConfig.Encoding = "hex"
		}
		if req.AuthConfig.Algorithm != "sha256" && req.AuthConfig.Algorithm != "sha512" {
			return errors.New("HMAC algorithm must be sha256 or sha512")
		}
		if req.AuthConfig.Encoding != "hex" && req.AuthConfig.Encoding != "base64" {
			return errors.New("HMAC encoding must be hex or base64")
		}
		return nil
	case app.TriggerAuthTypeAPIKey:
		if req.AuthConfig.Header == "" {
			req.AuthConfig.Header = "X-Nuon-API-Key"
		}
		return nil
	case app.TriggerAuthTypeBasic:
		if req.AuthConfig.Username == "" {
			req.AuthConfig.Username = "nuon"
		}
		return nil
	case app.TriggerAuthTypeBearerJWT:
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
		if req.AuthConfig.ExpectedEmail == "" && req.AuthConfig.ExpectedSubject == "" {
			return errors.New("bearer_jwt auth requires expected_email or expected_subject")
		}
		return nil
	default:
		return errors.New("unsupported auth_type")
	}
}

func applyTriggerPreset(req *createTriggerRequest) error {
	if req.Preset == "" {
		return nil
	}
	type preset struct {
		auth       app.TriggerAuthType
		envelope   app.EventEnvelopeType
		authConfig app.TriggerAuthConfig
		typeFrom   app.EventFieldSelector
		idFrom     app.EventFieldSelector
		google     bool
		sns        bool
	}
	presets := map[string]preset{
		"github":           {auth: app.TriggerAuthTypeHMAC, authConfig: app.TriggerAuthConfig{Header: "X-Hub-Signature-256", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"}, typeFrom: app.EventFieldSelector{Header: "X-GitHub-Event"}, idFrom: app.EventFieldSelector{Header: "X-GitHub-Delivery"}},
		"gitlab":           {auth: app.TriggerAuthTypeAPIKey, authConfig: app.TriggerAuthConfig{Header: "X-Gitlab-Token"}, typeFrom: app.EventFieldSelector{Header: "X-Gitlab-Event"}, idFrom: app.EventFieldSelector{Header: "X-Gitlab-Event-UUID"}},
		"bitbucket":        {auth: app.TriggerAuthTypeHMAC, authConfig: app.TriggerAuthConfig{Header: "X-Hub-Signature", Prefix: "sha256=", Algorithm: "sha256", Encoding: "hex"}, typeFrom: app.EventFieldSelector{Header: "X-Event-Key"}, idFrom: app.EventFieldSelector{Header: "X-Request-UUID"}},
		"gitea":            {auth: app.TriggerAuthTypeHMAC, authConfig: app.TriggerAuthConfig{Header: "X-Gitea-Signature", Algorithm: "sha256", Encoding: "hex"}, typeFrom: app.EventFieldSelector{Header: "X-Gitea-Event"}, idFrom: app.EventFieldSelector{Header: "X-Gitea-Delivery"}},
		"forgejo":          {auth: app.TriggerAuthTypeHMAC, authConfig: app.TriggerAuthConfig{Header: "X-Forgejo-Signature", Algorithm: "sha256", Encoding: "hex"}, typeFrom: app.EventFieldSelector{Header: "X-Forgejo-Event"}, idFrom: app.EventFieldSelector{Header: "X-Forgejo-Delivery"}},
		"terraform-cloud":  {auth: app.TriggerAuthTypeHMAC, authConfig: app.TriggerAuthConfig{Header: "X-TFE-Notification-Signature", Algorithm: "sha512", Encoding: "hex"}},
		"google-pubsub":    {auth: app.TriggerAuthTypeBearerJWT, envelope: app.EventEnvelopeTypePubSubPush, authConfig: app.TriggerAuthConfig{Issuer: "https://accounts.google.com"}, google: true},
		"azure-devops":     {auth: app.TriggerAuthTypeBasic, authConfig: app.TriggerAuthConfig{Username: "nuon"}},
		"aws-eventbridge":  {auth: app.TriggerAuthTypeAPIKey, authConfig: app.TriggerAuthConfig{Header: "X-Nuon-API-Key"}, typeFrom: app.EventFieldSelector{Payload: `$['detail-type']`}, idFrom: app.EventFieldSelector{Payload: "$.id"}},
		"aws-sns":          {auth: app.TriggerAuthTypeSNSSignature, envelope: app.EventEnvelopeTypeSNS, sns: true},
		"azure-event-grid": {auth: app.TriggerAuthTypeAPIKey, authConfig: app.TriggerAuthConfig{Header: "X-Nuon-API-Key"}},
		"slack-events":     {auth: app.TriggerAuthTypeHMAC, authConfig: app.TriggerAuthConfig{Header: "X-Slack-Signature", Prefix: "v0=", Algorithm: "sha256", Encoding: "hex"}},
		"datadog":          {auth: app.TriggerAuthTypeAPIKey, authConfig: app.TriggerAuthConfig{Header: "X-Nuon-API-Key"}, typeFrom: app.EventFieldSelector{Payload: "$.event_type"}, idFrom: app.EventFieldSelector{Payload: "$.event_id"}},
	}
	p, ok := presets[req.Preset]
	if !ok {
		return fmt.Errorf("unknown trigger preset %q", req.Preset)
	}
	if p.envelope == "" {
		p.envelope = app.EventEnvelopeTypeNone
	}
	if req.AuthType != "" && req.AuthType != p.auth {
		return fmt.Errorf("%s preset requires %s auth", req.Preset, p.auth)
	}
	if req.Envelope != "" && req.Envelope != p.envelope {
		return fmt.Errorf("%s preset requires the %s envelope", req.Preset, p.envelope)
	}
	desiredConfig := p.authConfig
	if p.google {
		desiredConfig.Audience = req.AuthConfig.Audience
		desiredConfig.ExpectedEmail = req.AuthConfig.ExpectedEmail
		desiredConfig.ExpectedSubject = req.AuthConfig.ExpectedSubject
	}
	if p.sns {
		desiredConfig.TopicARN = req.AuthConfig.TopicARN
	}
	providedConfig := req.AuthConfig
	if providedConfig.Header == "" {
		providedConfig.Header = desiredConfig.Header
	}
	if providedConfig.Prefix == "" {
		providedConfig.Prefix = desiredConfig.Prefix
	}
	if providedConfig.Encoding == "" {
		providedConfig.Encoding = desiredConfig.Encoding
	}
	if providedConfig.Algorithm == "" {
		providedConfig.Algorithm = desiredConfig.Algorithm
	}
	if providedConfig.Username == "" {
		providedConfig.Username = desiredConfig.Username
	}
	if providedConfig.Issuer == "" {
		providedConfig.Issuer = desiredConfig.Issuer
	}
	if !reflect.DeepEqual(providedConfig, desiredConfig) {
		return fmt.Errorf("%s preset has conflicting auth_config", req.Preset)
	}
	if req.TypeFrom != (app.EventFieldSelector{}) && req.TypeFrom != p.typeFrom {
		return fmt.Errorf("%s preset does not accept custom type_from", req.Preset)
	}
	if req.IDFrom != (app.EventFieldSelector{}) && req.IDFrom != p.idFrom {
		return fmt.Errorf("%s preset does not accept custom id_from", req.Preset)
	}
	req.AuthType, req.Envelope, req.AuthConfig = p.auth, p.envelope, desiredConfig
	req.TypeFrom, req.IDFrom = p.typeFrom, p.idFrom
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

// @ID ListTriggers
// @Summary List triggers
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Success 200 {array} triggerResponse
// @Router /v1/triggers [get]
func (s *service) ListTriggers(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}
	var triggers []app.Trigger
	if err := s.db.WithContext(ctx).Where(app.Trigger{OrgID: org.ID}).Find(&triggers).Error; err != nil {
		ctx.Error(err)
		return
	}
	secretsByTrigger := make(map[string][]app.TriggerSecret, len(triggers))
	if len(triggers) != 0 {
		triggerIDs := make([]string, 0, len(triggers))
		for i := range triggers {
			triggerIDs = append(triggerIDs, triggers[i].ID)
		}
		var secrets []app.TriggerSecret
		if err := s.db.WithContext(ctx).Where(map[string]any{"trigger_id": triggerIDs}).Order("created_at DESC").Find(&secrets).Error; err != nil {
			ctx.Error(err)
			return
		}
		for i := range secrets {
			triggerID := secrets[i].TriggerID
			if len(secretsByTrigger[triggerID]) < maxEventSecretMetadata {
				secretsByTrigger[triggerID] = append(secretsByTrigger[triggerID], secrets[i])
			}
		}
	}
	resp := make([]triggerResponse, 0, len(triggers))
	for i := range triggers {
		triggers[i].Secrets = secretsByTrigger[triggers[i].ID]
		resp = append(resp, triggerMetadataResponse(&triggers[i]))
	}
	ctx.JSON(http.StatusOK, resp)
}

// @ID GetTrigger
// @Summary Get an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {object} triggerResponse
// @Router /v1/triggers/{trigger_id} [get]
func (s *service) GetTrigger(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if ok {
		ctx.JSON(http.StatusOK, triggerMetadataResponse(trigger))
	}
}

// @ID GetTriggerIngressURL
// @Summary Retrieve the current trigger ingress URL
// @Description Uses PATCH so retrieving the sensitive ingress URL requires update permission instead of read permission.
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {object} triggerIngressURLResponse
// @Router /v1/triggers/{trigger_id}/ingress-url [patch]
func (s *service) GetTriggerIngressURL(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	ctx.JSON(http.StatusOK, triggerIngressURLResponse{IngressURL: buildIngressURL(s.cfg.PublicAPIURL, trigger.IngressKey)})
}

// @ID RotateTriggerIngressURL
// @Summary Replace an trigger ingress URL
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {object} credentialResponse
// @Router /v1/triggers/{trigger_id}/rotate-ingress-url [post]
func (s *service) RotateIngressURL(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	ingressKey, err := generateCredential()
	if err != nil {
		ctx.Error(err)
		return
	}
	updates := map[string]any{"ingress_key": ingressKey, "ingress_key_hash": hashIngressKey(ingressKey)}
	if err := s.db.WithContext(ctx).Model(trigger).Updates(updates).Error; err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusOK, credentialResponse{Trigger: *trigger, IngressURL: buildIngressURL(s.cfg.PublicAPIURL, ingressKey)})
}

// @ID DeleteTrigger
// @Summary Delete an trigger
// @Tags triggers
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Param force query bool false "Delete a trigger referenced by rules"
// @Success 204
// @Router /v1/triggers/{trigger_id} [delete]
func (s *service) DeleteTrigger(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	force := ctx.Query("force") == "true"
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var locked app.Trigger
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Trigger{ID: trigger.ID, OrgID: trigger.OrgID}).First(&locked).Error; err != nil {
			return err
		}
		var references int64
		if err := tx.Model(&app.TriggerRule{}).
			Where(app.TriggerRule{OrgID: locked.OrgID, TriggerID: locked.ID}).
			Count(&references).Error; err != nil {
			return err
		}
		if references != 0 && !force {
			return errTriggerReferenced
		}
		if err := tx.Model(&app.EventRunbookWaiter{}).
			Where(app.EventRunbookWaiter{OrgID: locked.OrgID, TriggerID: locked.ID, Status: app.EventRunbookWaiterStatusActive}).
			Count(&references).Error; err != nil {
			return err
		}
		if references != 0 {
			return errTriggerReferenced
		}
		now := time.Now()
		if err := tx.Unscoped().Model(&app.TriggerSecret{}).
			Where(app.TriggerSecret{TriggerID: locked.ID, OrgID: locked.OrgID}).
			Updates(map[string]any{"revoked_at": now, "secret": ""}).Error; err != nil {
			return err
		}
		return tx.Delete(&locked).Error
	})
	if errors.Is(err, errTriggerReferenced) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "trigger is referenced by trigger rules; use force=true to delete it"})
		return
	}
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (s *service) managementTrigger(ctx *gin.Context) (*app.Trigger, bool) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	trigger, err := s.scopedTrigger(ctx, org.ID, ctx.Param("trigger_id"))
	if err != nil {
		ctx.Error(err)
		return nil, false
	}
	return trigger, true
}

// @ID RotateTriggerSecret
// @Summary Rotate an trigger secret
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 201 {object} credentialResponse
// @Router /v1/triggers/{trigger_id}/rotate-secret [post]
func (s *service) RotateSecret(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	if trigger.Preset == "slack-events" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "update the Slack signing secret by recreating this trigger"})
		return
	}
	if !authUsesSecret(trigger.AuthType) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "trigger does not use a managed secret"})
		return
	}
	value, err := generateCredential()
	if err != nil {
		ctx.Error(err)
		return
	}
	now := time.Now()
	grace := now.Add(eventSecretRotationGrace)
	secret := app.TriggerSecret{OrgID: trigger.OrgID, TriggerID: trigger.ID, Secret: value, NotBefore: now}
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var lockedTrigger app.Trigger
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(app.Trigger{ID: trigger.ID, OrgID: trigger.OrgID}).First(&lockedTrigger).Error; err != nil {
			return err
		}
		if err := s.scrubInactiveTriggerSecrets(ctx, tx, trigger.ID, now); err != nil {
			return err
		}
		var active []app.TriggerSecret
		if err := tx.Where(app.TriggerSecret{TriggerID: trigger.ID}).Where(clause.Eq{Column: "revoked_at", Value: nil}).Where(clause.Or(clause.Eq{Column: "expires_at", Value: nil}, clause.Gt{Column: "expires_at", Value: now})).Order("created_at DESC").Find(&active).Error; err != nil {
			return err
		}
		for i := range active {
			previous := &active[i]
			if i == 0 {
				if previous.ExpiresAt == nil || previous.ExpiresAt.After(grace) {
					if err := tx.Model(previous).Update("expires_at", grace).Error; err != nil {
						return err
					}
				}
				continue
			}
			if err := tx.Model(previous).Updates(map[string]any{"revoked_at": now, "secret": ""}).Error; err != nil {
				return err
			}
		}
		return tx.Create(&secret).Error
	})
	if err != nil {
		ctx.Error(err)
		return
	}
	ctx.JSON(http.StatusCreated, credentialResponse{Trigger: *trigger, KeyID: secret.KeyID, Secret: value})
}

func (s *service) setTriggerStatus(ctx *gin.Context, status app.TriggerStatus) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	if err := s.db.WithContext(ctx).Model(trigger).Update("status", status).Error; err != nil {
		ctx.Error(err)
		return
	}
	trigger.Status = status
	ctx.JSON(http.StatusOK, trigger)
}

// @ID EnableTrigger
// @Summary Enable an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {object} app.Trigger
// @Router /v1/triggers/{trigger_id}/enable [post]
func (s *service) EnableTrigger(ctx *gin.Context) {
	s.setTriggerStatus(ctx, app.TriggerStatusActive)
}

// @ID DisableTrigger
// @Summary Disable an trigger
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Success 200 {object} app.Trigger
// @Router /v1/triggers/{trigger_id}/disable [post]
func (s *service) DisableTrigger(ctx *gin.Context) {
	s.setTriggerStatus(ctx, app.TriggerStatusSuspended)
}

// @ID RevokeTriggerSecret
// @Summary Revoke an trigger secret
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Param secret_id path string true "Trigger secret ID"
// @Success 200 {object} revokeTriggerSecretResponse
// @Router /v1/triggers/{trigger_id}/secrets/{secret_id}/revoke [post]
func (s *service) RevokeSecret(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	if trigger.Preset == "slack-events" {
		ctx.JSON(http.StatusConflict, gin.H{"error": "disable or delete this trigger to stop Slack events"})
		return
	}
	now := time.Now()
	res := s.db.WithContext(ctx).Model(&app.TriggerSecret{}).Where(app.TriggerSecret{ID: ctx.Param("secret_id"), OrgID: trigger.OrgID, TriggerID: trigger.ID}).Updates(map[string]any{"revoked_at": now, "secret": ""})
	if res.Error != nil {
		ctx.Error(res.Error)
		return
	}
	if res.RowsAffected == 0 {
		ctx.Error(gorm.ErrRecordNotFound)
		return
	}
	ctx.JSON(http.StatusOK, revokeTriggerSecretResponse{RevokedAt: now})
}

func secretIsRevealable(secret *app.TriggerSecret, now time.Time) bool {
	return secret.Secret != "" &&
		secret.RevokedAt == nil &&
		(secret.ExpiresAt == nil || secret.ExpiresAt.After(now))
}

// @ID RevealTriggerSecret
// @Summary Retrieve an active trigger secret value
// @Description Uses PATCH so retrieving the sensitive secret value requires update permission instead of read permission.
// @Tags triggers
// @Produce json
// @Security APIKey
// @Security OrgID
// @Param trigger_id path string true "Trigger ID"
// @Param secret_id path string true "Trigger secret ID"
// @Success 200 {object} revealTriggerSecretResponse
// @Router /v1/triggers/{trigger_id}/secrets/{secret_id}/reveal [patch]
func (s *service) RevealSecret(ctx *gin.Context) {
	trigger, ok := s.managementTrigger(ctx)
	if !ok {
		return
	}
	if trigger.Preset == "slack-events" {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Slack signing secrets are write-only"})
		return
	}
	var secret app.TriggerSecret
	if err := s.db.WithContext(ctx).Where(app.TriggerSecret{ID: ctx.Param("secret_id"), OrgID: trigger.OrgID, TriggerID: trigger.ID}).First(&secret).Error; err != nil {
		ctx.Error(err)
		return
	}
	if !secretIsRevealable(&secret, time.Now()) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "secret is no longer active"})
		return
	}
	ctx.JSON(http.StatusOK, revealTriggerSecretResponse{KeyID: secret.KeyID, Secret: secret.Secret})
}
