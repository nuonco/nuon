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

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type CurrentOrgWebhookResponse struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	WebhookURL  string    `json:"webhook_url"`
	CreatedByID string    `json:"created_by_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	HasSecret   bool      `json:"has_secret"`
}

type CreateCurrentOrgWebhookRequest struct {
	WebhookURL    string `json:"webhook_url" binding:"required"`
	WebhookSecret string `json:"webhook_secret"`
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

	webhook, err := s.createCurrentOrgWebhook(ctx, org.ID, account.ID, &req)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusCreated, toCurrentOrgWebhookResponse(*webhook))
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

	var existing app.Webhook
	err = s.db.WithContext(ctx).
		Where("org_id = ? AND webhook_url = ?", orgID, normalizedWebhookURL).
		First(&existing).Error
	if err == nil {
		return nil, stderr.ErrConflict{
			Err:         fmt.Errorf("webhook %q already exists for org %q", normalizedWebhookURL, orgID),
			Description: "webhook already exists for this org",
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to verify existing webhooks: %w", err)
	}

	webhook := app.Webhook{
		OrgID:         orgID,
		CreatedByID:   accountID,
		WebhookURL:    normalizedWebhookURL,
		WebhookSecret: normalizedWebhookSecret,
	}

	if err := s.db.WithContext(ctx).Create(&webhook).Error; err != nil {
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
	return CurrentOrgWebhookResponse{
		ID:          webhook.ID,
		OrgID:       webhook.OrgID,
		WebhookURL:  webhook.WebhookURL,
		CreatedByID: webhook.CreatedByID,
		CreatedAt:   webhook.CreatedAt,
		UpdatedAt:   webhook.UpdatedAt,
		HasSecret:   strings.TrimSpace(webhook.WebhookSecret) != "",
	}
}

func toCurrentOrgWebhookResponses(webhooks []app.Webhook) []CurrentOrgWebhookResponse {
	responses := make([]CurrentOrgWebhookResponse, 0, len(webhooks))
	for _, webhook := range webhooks {
		responses = append(responses, toCurrentOrgWebhookResponse(webhook))
	}

	return responses
}
