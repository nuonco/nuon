package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @BasePath /v1/vcs

// GetOrgConnections returns all VCS connections for an org
// @Summary get vcs connection for an org
// @Schemes
// @Description get vcs connections
// @Param org_id path string true "org ID for your current org"
// @Tags vcs
// @Accept json
// @Produce json
// @Success 200 {array} app.VCSConnection
// @Router /v1/vcs/{org_id}/connections [get]
func (s *service) GetOrgConnections(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	vcsConns, err := s.getOrgConnections(ctx, orgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org vcs connections: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, vcsConns)
}

func (s *service) getOrgConnections(ctx context.Context, orgID string) ([]*app.VCSConnection, error) {
	var vcsConns []*app.VCSConnection

	res := s.db.WithContext(ctx).Where("org_id = ?", orgID).Find(&vcsConns)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get vcs connections: %w", res.Error)
	}

	return vcsConns, nil
}
