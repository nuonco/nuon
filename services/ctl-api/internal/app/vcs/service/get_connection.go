package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
)

// @ID GetVCSConnection
// @Summary	returns a vcs connection for an org
// @Description.markdown	get_vcs_connection.md
// @Param			connection_id	path	string	true	"connection ID"
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
// @Success		200				{object}	app.VCSConnection
// @Router			/v1/vcs/connections/{connection_id} [get]
func (s *service) GetConnection(ctx *gin.Context) {
	vcsID := ctx.Param("connection_id")

	currentOrg, err := middlewares.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	vcsConn, err := s.getConnection(ctx, currentOrg.ID, vcsID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connection: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, vcsConn)
}

func (s *service) getConnection(ctx context.Context, orgID, vcsID string) (*app.VCSConnection, error) {
	vcsConn := app.VCSConnection{}

	res := s.db.WithContext(ctx).Where("org_id = ?", orgID).First(&vcsConn, "id = ?", vcsID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get vcs connection: %w", res.Error)
	}

	return &vcsConn, nil
}
