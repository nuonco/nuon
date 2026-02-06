package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type SetSlackWebhookURLRequest struct {
	Name *string `json:"name" validate:"required"`
}

// @ID						AdminSetInternalSlackWebhookURLOrg
// @Summary				set an internal slack webhook url for an org
// @Description.markdown	admin_set_org_slack_webhook_url.md
// @Param					org_id	path	string	true	"org ID for org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req	body	SetSlackWebhookURLRequest	true	"Input"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-internal-slack-webhook-url [POST]
func (s *service) AdminSetInternalSlackWebhookURLOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req SetSlackWebhookURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %v", err),
		})
		return
	}

	// Validate that name field was provided (but allow empty string)
	if err := s.v.Struct(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("validation failed: %w", err),
			Description: fmt.Sprintf("validation failed: %v", err),
		})
		return
	}

	webhookURL := ""
	if req.Name != nil {
		webhookURL = *req.Name
	}

	if err := s.setInternalOrgSlackWebhookURL(ctx, orgID, webhookURL); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) setInternalOrgSlackWebhookURL(ctx context.Context, orgID string, webhookURL string) error {
	// Use Select to force update even if value is empty string
	res := s.db.WithContext(ctx).
		Model(&app.NotificationsConfig{}).
		Where(&app.NotificationsConfig{
			OwnerID: orgID,
		}).
		Select("internal_slack_webhook_url").
		Updates(app.NotificationsConfig{
			InternalSlackWebhookURL: webhookURL,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update slack webhook url: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return stderr.ErrNotFound{
			Err:         fmt.Errorf("org notifications config not found"),
			Description: fmt.Sprintf("notifications config for org %s not found", orgID),
		}
	}

	return nil
}
