package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
)

type AdminAddVCSConnectionRequest struct {
	GithubInstallID string
}

// @ID						AdminAddVCSConnection
// @Summary				add a VCS connection for an org
// @Description.markdown	admin_add_vcs_connection.md
// @Tags					orgs/admin
// @Security				AdminEmail
// @Accept					json
// @Param					req		body	AdminAddVCSConnectionRequest	true	"Input"
// @Param					org_id	path	string							true	"org ID or name"
// @Produce				json
// @Success				201	{string}	ok
// @Router					/v1/orgs/{org_id}/admin-add-vcs-connection [POST]
func (s *service) AdminAddVCSConnection(ctx *gin.Context) {
	nameOrID := ctx.Param("org_id")

	var req AdminAddVCSConnectionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}

	org, err := s.adminGetOrg(ctx, nameOrID)
	if err != nil {
		ctx.Error(err)
		return
	}

	cctx.SetOrgGinContext(ctx, org)
	cctx.SetAccountIDGinContext(ctx, org.CreatedByID)

	if _, err := s.createOrgConnection(ctx, org.ID, req.GithubInstallID); err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, org)
}

func (s *service) createOrgConnection(ctx context.Context, orgID, githubInstallID string) (*app.VCSConnection, error) {
	vcsConn := app.VCSConnection{
		OrgID:           orgID,
		GithubInstallID: githubInstallID,
	}

	if err := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "github_install_id"}},
		DoNothing: true,
	}).Create(&vcsConn).Error; err != nil {
		return nil, fmt.Errorf("unable to create vcs_connection: %w", err)
	}

	// NOTE(jm): when this is a duplicate, the returned ID is not actually valid, as it is set by the create hook in
	// GORM, but then the conflict happens after.
	return &vcsConn, nil
}
