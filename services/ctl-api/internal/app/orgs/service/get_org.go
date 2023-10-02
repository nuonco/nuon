package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	orgmiddleware "github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

//	@BasePath	/v1/orgs

// Get an org
//
//	@Summary	Get an org
//	@Schemes
//	@Description	get an org
//	@Tags			orgs
//	@Accept			json
//	@Produce		json
//	@Param			X-Nuon-Org-ID	header		string	true	"org ID"
//	@Param			Authorization	header		string	true	"bearer auth token"
//	@Failure		400				{object}	stderr.ErrResponse
//	@Failure		404				{object}	stderr.ErrResponse
//	@Failure		500				{object}	stderr.ErrResponse
//	@Success		200				{object}	app.Org
//	@Router			/v1/orgs/current [GET]
func (s *service) GetOrg(ctx *gin.Context) {
	org, err := orgmiddleware.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	org, err = s.getOrg(ctx, org.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		Preload("UserOrgs").
		Preload("VCSConnections").
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return &org, nil
}
