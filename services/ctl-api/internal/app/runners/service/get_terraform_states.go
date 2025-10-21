package service

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/scopes"
)

// @ID						GetTerraformStatesV2
// @Summary				get terraform states
// @Description.markdown	get_terraform_states.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspaceState
// @Router					/v1/terraform-workspace/{workspace_id}/states [get]
func (s *service) GetTerraformWorkspaceStatesV2(ctx *gin.Context) {
	s.GetTerraformWorkspaceStates(ctx)
}

// @ID						GetTerraformStates
// @Summary				get terraform states
// @Description.markdown	get_terraform_states.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					runners,runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated  			true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspaceState
// @Router					/v1/runners/terraform-workspace/{workspace_id}/states [get]
func (s *service) GetTerraformWorkspaceStates(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")

	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})

		return
	}

	states, err := s.GetTerraformStates(ctx, workspaceID)
	if err != nil {
		ctx.Error(err)
		return
	}

	paginatedStates, err := db.HandlePaginatedResponse(ctx, states)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, paginatedStates)
}

func (s *service) GetTerraformStates(ctx *gin.Context, workspaceID string) ([]app.TerraformWorkspaceState, error) {
	states := []app.TerraformWorkspaceState{}
	res := s.db.WithContext(ctx).Model(&app.TerraformWorkspaceState{}).
		Scopes(scopes.WithOffsetPagination).Where("terraform_workspace_id = ?", workspaceID).
		Scopes(runnerJobPreload).
		Select(
			"ID",
			"CreatedByID",
			"CreatedAt",
			"UpdatedAt",
			"OrgID",
			"TerraformWorkspaceID",
			"Revision",
			"RunnerJobID",
		).
		Find(&states)
	if res.Error != nil {
		return nil, res.Error
	}

	return states, nil
}
