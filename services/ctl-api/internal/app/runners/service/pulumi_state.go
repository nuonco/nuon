package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @ID						GetPulumiState
// @Summary				get current pulumi state
// @Description.markdown	get_pulumi_state.md
// @Tags					runners/runner
// @Accept					json
// @Produce				octet-stream
// @Security				APIKey
// @Security				OrgID
// @Param					workspace_id	path	string	true	"workspace ID"
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{string}	string
// @Success				204	{string}	string
// @Router					/v1/runners/pulumi-state/{workspace_id} [get]
func (s *service) GetPulumiState(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")

	if _, err := s.getWorkspace(ctx, workspaceID); err != nil {
		ctx.Error(fmt.Errorf("unable to get workspace: %w", err))
		return
	}

	contents, err := s.helpers.GetTerraformStateJSON(ctx, workspaceID)
	if err != nil {
		s.l.Error("unable to get pulumi state", zap.Error(err))
		ctx.Error(fmt.Errorf("unable to get pulumi state: %w", err))
		return
	}

	if len(contents) == 0 {
		ctx.Status(http.StatusNoContent)
		return
	}

	ctx.Data(http.StatusOK, "application/octet-stream", contents)
}

// @ID						UpdatePulumiState
// @Summary				update pulumi state
// @Description.markdown	update_pulumi_state.md
// @Tags					runners/runner
// @Accept					octet-stream
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					job_id			query	string	false	"runner job ID"
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	string
// @Router					/v1/runners/pulumi-state/{workspace_id} [post]
func (s *service) UpdatePulumiState(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")

	if _, err := s.getWorkspace(ctx, workspaceID); err != nil {
		ctx.Error(fmt.Errorf("unable to get workspace: %w", err))
		return
	}

	jobID := ctx.Query("job_id")
	var sJobID *string
	if jobID != "" {
		sJobID = &jobID
	}

	contents, err := ctx.GetRawData()
	if err != nil {
		s.l.Error("unable to read request body", zap.Error(err))
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}

	if err := s.helpers.CreateStateJSON(ctx, workspaceID, sJobID, contents); err != nil {
		s.l.Error("unable to insert pulumi state", zap.Error(err))
		ctx.Error(fmt.Errorf("unable to update pulumi state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, "")
}
