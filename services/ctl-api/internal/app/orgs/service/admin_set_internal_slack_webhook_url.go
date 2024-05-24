package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SetSlackWebhookURLRequest struct {
	Name string `validate:"required"`
}

// @ID AdminSetSlackWebhookURLOrg
// @Summary set a slack webhook url for an org
// @Description.markdown admin_set_org_slack_webhook_url.md
// @Param			org_id	path	string	true	"org ID for org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	SetSlackWebhookURLRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-slack-webhook-url [POST]
func (s *service) AdminSetSlackWebhookURLOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req SetSlackWebhookURLRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.setOrgSlackWebhookURL(ctx, orgID, req.Name); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) setOrgSlackWebhookURL(ctx context.Context, orgID string, webhookURL string) error {
	res := s.db.WithContext(ctx).
		Where(&app.NotificationsConfig{
			OrgID: orgID,
		}).Updates(app.NotificationsConfig{
		InternalSlackWebhookURL: webhookURL,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update slack webhook url: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org notifications config not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
