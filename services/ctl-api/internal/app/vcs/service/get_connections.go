package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/org"
)

// @ID GetOrgVCSConnections
// @Summary	get vcs connection for an org
// @Description.markdown	get_org_vcs_connections.md
// @Tags			vcs
// @Accept			json
// @Produce		json
// @Security APIKey
// @Security OrgID
// @Failure		400				{object}	stderr.ErrResponse
// @Failure		401				{object}	stderr.ErrResponse
// @Failure		403				{object}	stderr.ErrResponse
// @Failure		404				{object}	stderr.ErrResponse
// @Failure		500				{object}	stderr.ErrResponse
// @Success		200				{array}		app.VCSConnection
// @Router			/v1/vcs/connections [get]
func (s *service) GetConnections(ctx *gin.Context) {
	currentOrg, err := org.FromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConns, err := s.getOrgConnections(ctx, currentOrg.ID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connections: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, vcsConns)
}

func (s *service) getOrgConnections(ctx context.Context, orgID string) ([]*app.VCSConnection, error) {
	var vcsConns []*app.VCSConnection

	res := s.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&vcsConns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get vcs connections: %w", res.Error)
	}

	return vcsConns, nil
}
