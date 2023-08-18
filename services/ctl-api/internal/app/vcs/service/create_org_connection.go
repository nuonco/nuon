package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm/clause"
)

type CreateOrgConnectionRequest struct {
	GithubInstallID string `json:"github_install_id" `
}

func (c *CreateOrgConnectionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

// @BasePath /v1/vcs

// PingExample godoc
// @Summary create a vcs connection for Github
// @Schemes
// @Description create a vcs connection
// @Param org_id path string true "org ID for your current org"
// @Param req body CreateOrgConnectionRequest true "Input"
// @Tags vcs
// @Accept json
// @Produce json
// @Success 200 {object} app.VCSConnection
// @Router /v1/vcs/{org_id}/connection [post]
func (s *service) CreateOrgConnection(ctx *gin.Context) {
	orgID := ctx.Param("org_id")

	var req CreateOrgConnectionRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
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
