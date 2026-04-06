package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type CreateConnectionRequest struct {
	GithubInstallID string `json:"github_install_id" validate:"required"`
}

func (c *CreateConnectionRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateVCSConnection
// @Summary					create a vcs connection for Github
// @Description.markdown	create_vcs_connection.md
// @Param					req	body	CreateConnectionRequest	true	"Input"
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	app.VCSConnection
// @Router					/v1/vcs/connections [post]
func (s *service) CreateConnection(ctx *gin.Context) {
	currentOrg, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	var req CreateConnectionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Fetch org name
	ghAccount, err := s.ghClient.GetInstallationAccount(ctx, req.GithubInstallID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org name: %w", err))
		return
	}
	ghAccountID := strconv.FormatInt(ghAccount.GetID(), 10)
	ghAccountName := ghAccount.GetLogin()

	// Create object
	vcsConn, err := s.createOrgConnection(ctx, currentOrg.ID, req.GithubInstallID, ghAccountID, ghAccountName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org connection: %w", err))
		return
	}

	// Create background health check queue for the connection
	if _, err := s.helpers.CreateConnectionQueue(ctx, vcsConn); err != nil {
		s.l.Warn("unable to create vcs connection queue",
			zap.String("vcs_connection_id", vcsConn.ID),
			zap.Error(err),
		)
	}

	ctx.JSON(http.StatusCreated, vcsConn)
}

func (s *service) createOrgConnection(
	ctx context.Context, orgID, githubInstallID, githubAccountID, githubAccountName string) (*app.VCSConnection, error) {
	vcsConn := app.VCSConnection{
		OrgID:             orgID,
		GithubInstallID:   githubInstallID,
		GithubAccountID:   githubAccountID,
		GithubAccountName: githubAccountName,
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
