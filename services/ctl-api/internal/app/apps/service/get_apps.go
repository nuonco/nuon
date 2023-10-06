package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

//	@BasePath	/v1/apps
//
// Get all apps for the current org
//
//	@Summary	get all apps for the current org
//	@Schemes
//	@Description	get an app
//	@Tags			apps
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		401				{object}	stderr.ErrResponse
//	@Failure		403				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{array}		app.App
//	@Router			/v1/apps [get]
func (s *service) GetApps(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	apps, err := s.getApps(ctx, org.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get apps for %s: %w", org.ID, err))
		return
	}
	ctx.JSON(http.StatusOK, apps)
}

func (s *service) getApps(ctx context.Context, orgID string) ([]*app.App, error) {
	var apps []*app.App
	org := &app.Org{
		ID: orgID,
	}

	err := s.db.WithContext(ctx).Preload("SandboxRelease").Model(&org).Association("Apps").Find(&apps)
	if err != nil {
		return nil, fmt.Errorf("unable to get org apps: %w", err)
	}

	return apps, nil
}
