package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type AdminDeleteIntegrationOrgsRequest struct{}

// @ID AdminDeleteIntegrationOrgs
// @Summary delete leaked integration orgs
// @Description.markdown delete_org.md
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	AdminDeleteIntegrationOrgsRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/admin-delete-integrations [POST]
func (s *service) AdminDeleteIntegrationOrgs(ctx *gin.Context) {
	orgs, err := s.getIntegrationOrgs(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, org := range orgs {
		if err := s.hardDeleteOrg(ctx, org.ID); err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) hardDeleteOrg(ctx context.Context, orgID string) error {
	// delete apps
	res := s.db.WithContext(ctx).Unscoped().
		Where("org_id = ?", orgID).
		Delete(&app.App{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org apps: %w", res.Error)
	}

	res = s.db.WithContext(ctx).Unscoped().Delete(&app.Org{
		ID: orgID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (s *service) getIntegrationOrgs(ctx context.Context) ([]app.Org, error) {
	var orgs []app.Org
	res := s.db.WithContext(ctx).
		Joins("JOIN  accounts on orgs.created_by_id=accounts.id").
		Where("accounts.account_type = ?", "integration").
		Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get canary orgs: %w", res.Error)
	}

	return orgs, nil
}
