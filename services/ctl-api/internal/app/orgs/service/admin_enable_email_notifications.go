package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						AdminEnableEmailNotifications
// @Summary				enable email notifications for an org
// @Param					org_id	path	string	true	"org ID for org"
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				200	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-enable-email-notifications [POST]
func (s *service) AdminEnableEmailNotifications(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	_, err := s.adminGetOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	res := s.db.WithContext(ctx).
		Model(&app.NotificationsConfig{}).
		Where(&app.NotificationsConfig{
			OwnerID: orgID,
		}).
		Update("enable_email_notifications", true)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to enable email notifications: %w", res.Error))
		return
	}
	if res.RowsAffected != 1 {
		ctx.Error(fmt.Errorf("org notifications config not found: %w", gorm.ErrRecordNotFound))
		return
	}

	ctx.JSON(http.StatusOK, true)
}
