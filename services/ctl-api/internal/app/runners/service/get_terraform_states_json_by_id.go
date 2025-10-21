package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/pkg/errors"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						GetTerraformWorkspaceStatesJSONByIDV2
// @Summary				get terraform state json by id. This output is same as "terraform show --json"
// @Description.markdown	get_terraform_states_json_by_id.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id					path	string	true	"terraform state ID"
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
// @Router					/v1/terraform-workspaces/{workspace_id}/state-json/{state_id} [get]
func (s *service) GetTerraformWorkspaceStatesJSONByIDV2(ctx *gin.Context) {
	s.GetTerraformWorkspaceStatesJSONByID(ctx)
}

// @ID						GetTerraformWorkspaceStatesJSONByID
// @Summary				get terraform state json by id. This output is same as "terraform show --json"
// @Description.markdown	get_terraform_states_json_by_id.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id					path	string	true	"terraform state ID"
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
// @Success				200	{object}	object{}
// @Router					/v1/runners/terraform-workspace/{workspace_id}/state-json/{state_id} [get]
func (s *service) GetTerraformWorkspaceStatesJSONByID(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	if workspaceID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id was not set"),
		})
		return
	}

	stateID := ctx.Param("state_id")
	if stateID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("state_id was not set"),
		})
		return
	}

	state, err := s.GetTerraformStatesJSONById(ctx, workspaceID, stateID)
	if err != nil {
		ctx.Error(err)
		return
	}

	sanitized := strings.Trim(string(state.Contents), "\"' \n\r\t")
	decodedReader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(sanitized))

	var builder strings.Builder
	if _, err := io.Copy(&builder, decodedReader); err != nil {
		ctx.Error(fmt.Errorf("unable to decode base64 string: %w", err))
		return
	}

	var statejson *tfjson.State
	err = json.Unmarshal([]byte(builder.String()), &statejson)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to unmarshal terraform state: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, statejson)

}

func (s *service) GetTerraformStatesJSONById(ctx *gin.Context, workspaceID string, stateID string) (*app.TerraformWorkspaceStateJSON, error) {
	state := &app.TerraformWorkspaceStateJSON{}
	res := s.db.WithContext(ctx).Model(&app.TerraformWorkspaceStateJSON{}).
		Where("id = ? and workspace_id = ?", stateID, workspaceID).First(state)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get terraform state: %w", res.Error)
	}

	return state, nil
}
