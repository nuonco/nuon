package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						UpdateTerraformStateJSON
// @Summary				update terraform state json
// @Description.markdown	update_terraform_state_json.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param job_id 				query	string	false	"job ID"
// @Param					body body interface{} true "terraform workspace unlock "
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	interface{}
// @Router					/v1/terraform-workspaces/{workspace_id}/state-json [post]
func (s *service) UpdateTerraformWorkspaceStateJSON(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	// keeping jobID optional to remain backwards compatible for old runners
	jobID := ctx.Query("job_id")
	var sJobID *string
	if jobID != "" {
		sJobID = &jobID
	}

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to read request body: %w", err))
		return
	}

	if err := s.helpers.UpdateStateJSON(ctx, workspaceID, sJobID, body); err != nil {
		ctx.Error(fmt.Errorf("unable to update state json: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "terraform state json updated"})
}
