package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

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
