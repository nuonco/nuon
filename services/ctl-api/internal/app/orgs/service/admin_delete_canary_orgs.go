package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type AdminDeleteCanaryOrgsRequest struct{}

// @ID AdminDeleteCanaryOrgs
// @Summary delete canary orgs
// @Description.markdown delete_org.md
// @Tags			orgs/admin
// @Accept			json
// @Param			req	body	AdminDeleteCanaryOrgsRequest	true	"Input"
// @Produce		json
// @Success		201	{string}	ok
// @Router			/v1/orgs/admin-delete-canarys [POST]
func (s *service) AdminDeleteCanaryOrgs(ctx *gin.Context) {
	orgs, err := s.getCanaryOrgs(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	for _, org := range orgs {
		for _, app := range org.Apps {
			for _, install := range app.Installs {
				s.installHooks.Forgotten(ctx, install.ID)
			}
			s.appHooks.Deleted(ctx, app.ID)
		}
		s.hooks.ForceDelete(ctx, org.ID)
	}

	ctx.JSON(http.StatusOK, true)
}

func (s *service) getCanaryOrgs(ctx context.Context) ([]app.Org, error) {
	var orgs []app.Org
	res := s.db.WithContext(ctx).
		Preload("Apps").
		Preload("Apps.Installs").
		Joins("JOIN user_tokens on orgs.created_by_id=user_tokens.subject").
		Where("user_tokens.token_type = ?", "canary").
		Find(&orgs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get canary orgs: %w", res.Error)
	}

	return orgs, nil
}
