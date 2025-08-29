package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetOrgVCSConnections
// @Summary				get vcs connection for an org
// @Description.markdown	get_org_vcs_connections.md
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					vcs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}		app.VCSConnection
// @Router					/v1/vcs/connections [get]
func (s *service) GetConnections(ctx *gin.Context) {
	currentOrg, err := cctx.OrgFromContext(ctx)
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

func (s *service) getOrgConnections(ctx *gin.Context, orgID string) ([]*app.VCSConnection, error) {
	var vcsConns []*app.VCSConnection

	res := s.db.
		Scopes(scopes.WithOffsetPagination).
		WithContext(ctx).Where("org_id = ?", orgID).Find(&vcsConns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get vcs connections: %w", res.Error)
	}

	vcsConns, err := db.HandlePaginatedResponse(ctx, vcsConns)
	if err != nil {
		return nil, fmt.Errorf("unable to handle paginated response: %w", err)
	}

	return vcsConns, nil
}
