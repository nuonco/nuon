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

// @ID						GetTerraformWorkspaceStatesJSONV2
// @Summary				get terraform states json
// @Description.markdown	get_terraform_states_json.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspaceStateJSON
// @Router					/v1/terraform-workspaces/{workspace_id}/state-json [get]
func (s *service) GetTerraformWorkspaceStatesJSONV2(ctx *gin.Context) {
	s.GetTerraformWorkspaceStatesJSON(ctx)
}

// @ID						GetTerraformWorkspaceStatesJSON
// @Summary				get terraform states json
// @Description.markdown	get_terraform_states_json.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					offset						query	int		false	"offset of results to return"	Default(0)
// @Param					limit						query	int		false	"limit of results to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Deprecated     true
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{array}	app.TerraformWorkspaceStateJSON
// @Router					/v1/runners/terraform-workspace/{workspace_id}/state-json [get]
func (s *service) GetTerraformWorkspaceStatesJSON(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	states, err := s.GetTerraformStatesJSON(ctx, workspaceID)
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

func (s *service) GetTerraformStatesJSON(ctx *gin.Context, workspaceID string) ([]app.TerraformWorkspaceStateJSON, error) {
	states := []app.TerraformWorkspaceStateJSON{}
	res := s.db.WithContext(ctx).Model(&app.TerraformWorkspaceStateJSON{}).
		Scopes(scopes.WithOffsetPagination).Where("workspace_id = ?", workspaceID).
		Select(
			"ID",
			"CreatedByID",
			"CreatedAt",
			"UpdatedAt",
			"WorkspaceID",
			"RunnerJobID",
		).
		Scopes(runnerJobPreload).
		Order("created_at DESC").
		Find(&states)
	if res.Error != nil {
		return nil, res.Error
	}

	return states, nil
}
