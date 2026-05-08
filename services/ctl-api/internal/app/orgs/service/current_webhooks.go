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

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/interests"
)

type CurrentOrgWebhookResponse struct {
	ID          string                    `json:"id"`
	OrgID       string                    `json:"org_id"`
	WebhookURL  string                    `json:"webhook_url"`
	CreatedByID string                    `json:"created_by_id"`
	CreatedAt   time.Time                 `json:"created_at"`
	UpdatedAt   time.Time                 `json:"updated_at"`
	HasSecret   bool                      `json:"has_secret"`
	Interests   interests.Interests       `json:"interests" swaggertype:"object"`
	Match       *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
}

type CreateCurrentOrgWebhookRequest struct {
	WebhookURL    string                    `json:"webhook_url" binding:"required"`
	WebhookSecret string                    `json:"webhook_secret"`
	Interests     interests.Interests       `json:"interests" swaggertype:"object"`
	Match         *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
}

// UpdateCurrentOrgWebhookRequest mutates a webhook's interests, scope (Match),
// and/or secret. WebhookURL is part of the (org_id, webhook_url, match)
// unique index and cannot be changed in place — callers should delete +
// create to rename. Interests is always replaced wholesale (PUT semantics
// for the field). Match is also replaced wholesale: passing nil resets the
// row to org-wide; passing a non-nil predicate replaces the existing scope.
// WebhookSecret pointer distinguishes "leave unchanged" (nil) from "clear"
// (empty string).
type UpdateCurrentOrgWebhookRequest struct {
	WebhookSecret *string                   `json:"webhook_secret,omitempty"`
	Interests     interests.Interests       `json:"interests" swaggertype:"object"`
	Match         *labels.SubscriptionMatch `json:"match,omitempty" swaggertype:"object"`
}

// @ID                                          GetCurrentOrgWebhooks
// @Summary                             list webhooks for the current org
// @Tags                                        orgs
// @Accept                                      json
// @Produce                             json
// @Security                            APIKey
// @Security                            OrgID
// @Success                             200     {array}         CurrentOrgWebhookResponse
// @Failure                             400     {object}        stderr.ErrResponse
// @Failure                             401     {object}        stderr.ErrResponse
// @Failure                             403     {object}        stderr.ErrResponse
// @Failure                             500     {object}        stderr.ErrResponse
// @Router                                      /v1/orgs/current/webhooks [GET]
func (s *service) GetCurrentOrgWebhooks(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	webhooks, err := s.listCurrentOrgWebhooks(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, toCurrentOrgWebhookResponses(webhooks))
}

// @ID                                          CreateCurrentOrgWebhook
// @Summary                             create a webhook for the current org
// @Tags                                        orgs
// @Accept                                      json
// @Produce                             json
// @Security                            APIKey
// @Security                            OrgID
// @Param                                       req             body            CreateCurrentOrgWebhookRequest        true    "Input"
// @Success                             201             {object}        CurrentOrgWebhookResponse
// @Failure                             400             {object}        stderr.ErrResponse
// @Failure                             401             {object}        stderr.ErrResponse
// @Failure                             403             {object}        stderr.ErrResponse
// @Failure                             409             {object}        stderr.ErrResponse
// @Failure                             500             {object}        stderr.ErrResponse
// @Router                                      /v1/orgs/current/webhooks [POST]
func (s *service) CreateCurrentOrgWebhook(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	account, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateCurrentOrgWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := interests.Validate(req.Interests); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if req.Match != nil {
		if err := req.Match.Validate(); err != nil {
			ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid match: %w", err)))
			return
		}
	}

	webhook, err := s.createCurrentOrgWebhook(ctx, org.ID, account.ID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, toCurrentOrgWebhookResponse(*webhook))
}

// @ID                                          UpdateCurrentOrgWebhook
// @Summary                             update a webhook for the current org
// @Description                 Replaces the webhook's interests filter and/or rotates its signing secret. WebhookURL is part of the (org_id, webhook_url) unique index and cannot be changed in place — delete and recreate to rename.
// @Tags                                        orgs
// @Accept                                      json
// @Produce                             json
// @Security                            APIKey
// @Security                            OrgID
// @Param                                       webhook_id      path            string                                              true    "webhook ID"
// @Param                                       req                     body            UpdateCurrentOrgWebhookRequest      true    "Input"
// @Success                             200             {object}        CurrentOrgWebhookResponse
// @Failure                             400             {object}        stderr.ErrResponse
// @Failure                             401             {object}        stderr.ErrResponse
// @Failure                             403             {object}        stderr.ErrResponse
// @Failure                             404             {object}        stderr.ErrResponse
// @Failure                             500             {object}        stderr.ErrResponse
// @Router                                      /v1/orgs/current/webhooks/{webhook_id} [PATCH]
func (s *service) UpdateCurrentOrgWebhook(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req UpdateCurrentOrgWebhookRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := interests.Validate(req.Interests); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if req.Match != nil {
		if err := req.Match.Validate(); err != nil {
			ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("invalid match: %w", err)))
			return
		}
	}

	webhook, err := s.updateCurrentOrgWebhook(ctx, org.ID, ctx.Param("webhook_id"), &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, toCurrentOrgWebhookResponse(*webhook))
}

func (s *service) updateCurrentOrgWebhook(ctx context.Context, orgID, webhookID string, req *UpdateCurrentOrgWebhookRequest) (*app.Webhook, error) {
	trimmedWebhookID := strings.TrimSpace(webhookID)
	if trimmedWebhookID == "" {
		return nil, stderr.ErrUser{
			Err:         fmt.Errorf("webhook_id is required"),
			Description: "webhook_id is required",
		}
	}

	var webhook app.Webhook
	res := s.db.WithContext(ctx).
		Where(app.Webhook{ID: trimmedWebhookID, OrgID: orgID}).
		First(&webhook)
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil, stderr.ErrNotFound{
			Err:         fmt.Errorf("webhook %q not found for org %q", trimmedWebhookID, orgID),
			Description: "webhook not found",
		}
	}
	if res.Error != nil {
		return nil, fmt.Errorf("unable to load webhook: %w", res.Error)
	}

	updates := map[string]any{
		"interests": req.Interests,
		// Match is replaced wholesale — passing nil resets to org-wide.
		// match_canonical is computed inline (rather than relying on
		// BeforeSave, which doesn't fire reliably for map-based Updates)
		// so the unique index always sees the canonical projection.
		"match":           req.Match,
		"match_canonical": req.Match.Canonical(),
	}
	if req.WebhookSecret != nil {
		updates["webhook_secret"] = strings.TrimSpace(*req.WebhookSecret)
	}

	if err := s.db.WithContext(ctx).
		Model(&webhook).
		Updates(updates).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, stderr.ErrConflict{
				Err:         fmt.Errorf("webhook %q for org %q already exists with this scope: %w", trimmedWebhookID, orgID, err),
				Description: "another webhook for this URL already uses this scope",
			}
		}
		return nil, fmt.Errorf("unable to update webhook: %w", err)
	}

	// Reflect the new values on the returned struct so the response body
	// surfaces the updated scope without an extra SELECT.
	webhook.Interests = req.Interests
	webhook.Match = req.Match
	webhook.MatchCanonical = req.Match.Canonical()
	if req.WebhookSecret != nil {
		webhook.WebhookSecret = strings.TrimSpace(*req.WebhookSecret)
	}

	return &webhook, nil
}

// isUniqueViolation sniffs a GORM error for the Postgres 23505 unique
// constraint code. Used to convert duplicate-Match upserts into 409s.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate key value") ||
		strings.Contains(msg, "SQLSTATE 23505")
}

// @ID                                          DeleteCurrentOrgWebhook
// @Summary                             delete a webhook for the current org
// @Tags                                        orgs
// @Accept                                      json
// @Produce                             json
// @Security                            APIKey
// @Security                            OrgID
// @Param                                       webhook_id      path            string  true    "webhook ID"
// @Success                             204
// @Failure                             400             {object}        stderr.ErrResponse
// @Failure                             401             {object}        stderr.ErrResponse
// @Failure                             403             {object}        stderr.ErrResponse
// @Failure                             404             {object}        stderr.ErrResponse
// @Failure                             500             {object}        stderr.ErrResponse
// @Router                                      /v1/orgs/current/webhooks/{webhook_id} [DELETE]
func (s *service) DeleteCurrentOrgWebhook(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if err := s.deleteCurrentOrgWebhook(ctx, org.ID, ctx.Param("webhook_id")); err != nil {
		ctx.Error(err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (s *service) listCurrentOrgWebhooks(ctx context.Context, orgID string) ([]app.Webhook, error) {
	var webhooks []app.Webhook
	if err := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Order("created_at DESC").
		Find(&webhooks).Error; err != nil {
		return nil, fmt.Errorf("unable to list webhooks: %w", err)
	}

	return webhooks, nil
}

func (s *service) createCurrentOrgWebhook(ctx context.Context, orgID string, accountID string, req *CreateCurrentOrgWebhookRequest) (*app.Webhook, error) {
	normalizedWebhookURL, err := normalizeWebhookURL(req.WebhookURL)
	if err != nil {
		return nil, stderr.ErrUser{
			Err:         err,
			Description: err.Error(),
		}
	}

	normalizedWebhookSecret := strings.TrimSpace(req.WebhookSecret)

	// Conflict check now keys on (org_id, webhook_url, match_canonical) to
	// match the unique index introduced by migration 102: the same URL can
	// be registered multiple times in the same org with different scopes.
	matchCanonical := req.Match.Canonical()
	var existing app.Webhook
	err = s.db.WithContext(ctx).
		Where("org_id = ? AND webhook_url = ? AND match_canonical = ?",
			orgID, normalizedWebhookURL, matchCanonical).
		First(&existing).Error
	if err == nil {
		return nil, stderr.ErrConflict{
			Err:         fmt.Errorf("webhook %q already exists for org %q with this scope", normalizedWebhookURL, orgID),
			Description: "webhook already exists for this org with this scope",
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to verify existing webhooks: %w", err)
	}

	// Default new webhooks to AllEvents=true so they receive every supported
	// event until the caller explicitly opts into a per-resource config.
	// Mirrors the slack channel subscription create handler.
	subInterests := req.Interests
	if subInterests.IsZero() {
		subInterests = interests.AllEvents()
	}

	webhook := app.Webhook{
		OrgID:         orgID,
		CreatedByID:   accountID,
		WebhookURL:    normalizedWebhookURL,
		WebhookSecret: normalizedWebhookSecret,
		Interests:     subInterests,
		Match:         req.Match,
	}

	if err := s.db.WithContext(ctx).Create(&webhook).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, stderr.ErrConflict{
				Err:         fmt.Errorf("webhook %q already exists for org %q with this scope: %w", normalizedWebhookURL, orgID, err),
				Description: "webhook already exists for this org with this scope",
			}
		}
		return nil, fmt.Errorf("unable to create webhook: %w", err)
	}

	return &webhook, nil
}

func (s *service) deleteCurrentOrgWebhook(ctx context.Context, orgID string, webhookID string) error {
	trimmedWebhookID := strings.TrimSpace(webhookID)
	if trimmedWebhookID == "" {
		return stderr.ErrUser{
			Err:         fmt.Errorf("webhook_id is required"),
			Description: "webhook_id is required",
		}
	}

	res := s.db.WithContext(ctx).
		Where("id = ? AND org_id = ?", trimmedWebhookID, orgID).
		Delete(&app.Webhook{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete webhook: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return stderr.ErrNotFound{
			Err:         fmt.Errorf("webhook %q not found for org %q", trimmedWebhookID, orgID),
			Description: "webhook not found",
		}
	}

	return nil
}

func normalizeWebhookURL(rawWebhookURL string) (string, error) {
	trimmedWebhookURL := strings.TrimSpace(rawWebhookURL)
	if trimmedWebhookURL == "" {
		return "", fmt.Errorf("webhook_url is required")
	}

	parsed, err := url.ParseRequestURI(trimmedWebhookURL)
	if err != nil {
		return "", fmt.Errorf("webhook_url must be a valid absolute URL: %w", err)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("webhook_url must include a host")
	}

	scheme := strings.ToLower(parsed.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", fmt.Errorf("webhook_url must use http or https scheme")
	}

	return trimmedWebhookURL, nil
}

func toCurrentOrgWebhookResponse(webhook app.Webhook) CurrentOrgWebhookResponse {
	// Surface the effective interests config: rows persisted before the
	// interests filter shipped store NULL/zero JSONB but are delivered as
	// AllEvents=true (see hooks/webhook.go listOrgWebhookTargets). Returning
	// the same shape here keeps the CLI/dashboard/SDK in sync with what's
	// actually delivered.
	effectiveInterests := webhook.Interests
	if effectiveInterests.IsZero() {
		effectiveInterests = interests.AllEvents()
	}

	return CurrentOrgWebhookResponse{
		ID:          webhook.ID,
		OrgID:       webhook.OrgID,
		WebhookURL:  webhook.WebhookURL,
		CreatedByID: webhook.CreatedByID,
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
		HasSecret:   strings.TrimSpace(webhook.WebhookSecret) != "",
		Interests:   effectiveInterests,
		Match:       webhook.Match,
	}
}

func toCurrentOrgWebhookResponses(webhooks []app.Webhook) []CurrentOrgWebhookResponse {
	responses := make([]CurrentOrgWebhookResponse, 0, len(webhooks))
	for _, webhook := range webhooks {
		responses = append(responses, toCurrentOrgWebhookResponse(webhook))
	}

	return responses
}
