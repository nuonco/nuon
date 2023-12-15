package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminDeleteOrgRequest struct{}

// Delete an org and everything in it
//
//	@Summary delete an org and everything in it
//
// @Schemes
//
//	@Description	delete an org org
//	@Param			org_id	path	string	true	"org ID for your current org"
//	@Tags			orgs/admin
//	@Accept			json
//	@Param			req	body	AdminDeleteOrgRequest	true	"Input"
//	@Produce		json
//	@Success		201	{string}	ok
//	@Router			/v1/orgs/{org_id}/admin-delete [POST]
func (s *service) AdminDeleteOrg(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	org, err := s.getOrgAndDependencies(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}
	orgID = org.ID

	err = s.deleteOrg(ctx, orgID)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, app := range org.Apps {
		s.appHooks.Deleted(ctx, app.ID)
		for _, install := range app.Installs {
			s.installHooks.Deleted(ctx, install.ID)
			s.installHooks.Forgotten(ctx, install.ID)
		}

		for _, comp := range app.Components {
			s.componentHooks.Deleted(ctx, comp.ID)
		}
	}
	s.hooks.Deleted(ctx, orgID)
	ctx.JSON(http.StatusOK, true)
}

func (s *service) getOrgAndDependencies(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Installs").
		Preload("Apps.Components").
		Where("name = ?", orgID).
		Or("id = ?", orgID).
		First(&org)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return &org, nil
}
