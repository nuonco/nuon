package service

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateConnectionCallbackRequest struct {
	OrgID           string `json:"org_id" validate:"required"`
	GithubInstallID string `json:"github_install_id" validate:"required"`
}

func (c *CreateConnectionCallbackRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						CreateVCSConnectionCallback
// @Summary					public connection to create a vcs connection via a callback
// @Description.markdown	create_connection_callback.md
// @Tags					vcs
// @Accept					json
// @Produce					json
// @Param					req	body		CreateConnectionCallbackRequest	true	"Input"
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					201	{object}	app.VCSConnection
// @Router					/v1/vcs/connection-callback [post]
func (s *service) CreateConnectionCallback(ctx *gin.Context) {
	var req CreateConnectionCallbackRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	org, err := s.getOrg(ctx, req.OrgID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org: %w", err))
		return
	}

	// Fetch org name
	ghAccount, err := s.getGithubAccount(ctx, req.GithubInstallID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org name: %w", err))
		return
	}
	ghAccountID := strconv.FormatInt(ghAccount.GetID(), 10)
	ghAccountName := ghAccount.GetLogin()

	// Create object
	dbCtx := cctx.SetAccountIDContext(ctx, org.CreatedByID)
	vcsConn, err := s.createOrgConnection(dbCtx, req.OrgID, req.GithubInstallID, ghAccountID, ghAccountName)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create org connection: %w", err))
		return
	}

	ctx.JSON(http.StatusCreated, vcsConn)
}

func (s *service) getOrg(ctx context.Context, orgID string) (*app.Org, error) {
	org := app.Org{}
	res := s.db.WithContext(ctx).
		First(&org, "id = ?", orgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org %s: %w", orgID, res.Error)
	}

	return &org, nil
}
