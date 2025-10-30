package service

import (
"fmt"
"net/http"

"github.com/gin-gonic/gin"
"github.com/go-playground/validator/v10"

"github.com/powertoolsdev/mono/services/ctl-api/internal/app/app-branches/signals"
"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/validator"
)

type CreateAppBranchRequest struct {
	Name                       string `json:"name" validate:"required,min=1"`
	ConnectedGithubVCSConfigID string `json:"connected_github_vcs_config_id" validate:"required"`
}

func (c *CreateAppBranchRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return validatorPkg.FormatValidationError(err)
	}

	return nil
}

// @ID						CreateAppBranch
// @Description.markdown	create_app_branch.md
// @Tags					apps
// @Accept					json
// @Param					req	body	CreateAppBranchRequest	true	"Input"
// @Produce				json
// @Param					app_id	path	string	true	"app ID"
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranch
// @Router					/v1/apps/{app_id}/branches [post]
func (s *service) CreateAppBranch(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	appID := ctx.Param("app_id")

	var req CreateAppBranchRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	branch, err := s.helpers.CreateAppBranch(ctx, org.ID, appID, req.Name, req.ConnectedGithubVCSConfigID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch: %w", err))
		return
	}

	s.evClient.Send(ctx, appID, &signals.Signal{
		Type:        signals.OperationCreated,
		AppBranchID: branch.ID,
	})

	ctx.JSON(http.StatusCreated, branch)
}
