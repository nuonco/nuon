package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type CreateOrgConnectionRequest struct {
	GithubInstallID string `json:"github_install_id"`
}

func (s *service) CreateOrgConnection(ctx *gin.Context) {
	orgID := ctx.Param("org_id")
	var req CreateOrgConnectionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	vcsConn, err := s.createOrgConnection(ctx, orgID, &req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org connection: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, vcsConn)
}

func (s *service) createOrgConnection(ctx context.Context, orgID string, req *CreateOrgConnectionRequest) (*app.VCSConnection, error) {
	vcsConn := app.VCSConnection{
		OrgID:           orgID,
		GithubInstallID: req.GithubInstallID,
	}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).Create(&vcsConn).Error; err != nil {
		return nil, fmt.Errorf("unable to create vcs_connection: %w", err)
	}

	// NOTE(jm): when this is a duplicate, the returned ID is not actually valid, as it is set by the create hook in
	// GORM, but then the conflict happens after.
	return &vcsConn, nil
}
