package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type SetSlackWebhookURLRequest struct {
	Name string `json:"name" validate:"required"`
}

func (r *SetSlackWebhookURLRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
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

	_, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req SetSlackWebhookURLRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("unable to parse request: %w", err),
			Description: fmt.Sprintf("unable to parse request: %s", err.Error()),
		})
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	if err := s.setInternalOrgSlackWebhookURL(ctx, orgID, req.Name); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) setInternalOrgSlackWebhookURL(ctx context.Context, orgID string, webhookURL string) error {
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
