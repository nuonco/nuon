package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type SetCustomerSlackWebhookURLRequest struct {
	Name string `validate:"required"`
}

// @ID AdminSetCustomerSlackWebhookURLOrg
// @Summary set a customer slack webhook url for an org
// @Description.markdown admin_set_org_slack_webhook_url.md
// @Param			org_id	path	string	true	"org ID for org"
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	SetCustomerSlackWebhookURLRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/{org_id}/admin-customer-slack-webhook-url [POST]
func (s *service) AdminSetCustomerSlackWebhookURLOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.getOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req SetCustomerSlackWebhookURLRequest
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
			OwnerID: orgID,
		}).Updates(app.NotificationsConfig{
		SlackWebhookURL: webhookURL,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to update slack webhook url: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org notifications config not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}
