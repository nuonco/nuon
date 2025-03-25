package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminDeleteIntegrationOrgsRequest struct{}

//	@ID						AdminDeleteIntegrationOrgs
//	@Summary				delete leaked integration orgs
//	@Description.markdown	delete_org.md
//	@Tags					orgs/admin
//	@Security				AdminEmail
//	@Accept					json
//	@Param					req	body	AdminDeleteIntegrationOrgsRequest	true	"Input"
//	@Produce				json
//	@Success				201	{string}	ok
//	@Router					/v1/orgs/admin-delete-integrations [POST]
func (s *service) AdminDeleteIntegrationOrgs(ctx *gin.Context) {
	orgs, err := s.getIntegrationOrgs(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, org := range orgs {
		if err := s.helpers.HardDelete(ctx, org.ID); err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.JSON(http.StatusOK, true)
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
