package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// @ID						GetWorkspaceStateJSONRawByID
// @Summary				get raw workspace state json by id
// @Description			Returns the raw state contents without format-specific parsing. Works for both terraform and pulumi state.
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id		path	string	true	"state ID"
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
// @Success				200	{object}	object{}
// @Router					/v1/terraform-workspaces/{workspace_id}/state-json/{state_id}/raw [get]
func (s *service) GetWorkspaceStateJSONRawByID(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{Err: fmt.Errorf("workspace_id is required")})
		return
	}

	stateID := ctx.Param("state_id")
	if stateID == "" {
		ctx.Error(stderr.ErrInvalidRequest{Err: fmt.Errorf("state_id is required")})
		return
	}

	if _, err := s.getWorkspace(ctx, workspaceID); err != nil {
		ctx.Error(fmt.Errorf("unable to get workspace: %w", err))
		return
	}

	state, err := s.GetTerraformStatesJSONById(ctx, workspaceID, stateID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Data(http.StatusOK, "application/json", state.Contents)
}
