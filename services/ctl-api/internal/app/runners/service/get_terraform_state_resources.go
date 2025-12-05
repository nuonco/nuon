package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

// @ID						GetTerraformWorkspaceStateJSONResourcesV2
// @Summary				get terraform state resources. This output is similar to "terraform state list"
// @Description.markdown	get_terraform_state_json_resources.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id 		path	string	true	"state ID"
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
// @Success				200	{object}	interface{}
// @Router					/v1/terraform-workspace/{workspace_id}/state-json/{state_id}/resources [get]
func (s *service) GetTerraformWorkspaceStateResourcesV2(ctx *gin.Context) {
	s.GetTerraformWorkspaceStateResources(ctx)
}

// @ID						GetTerraformWorkspaceStateJSONResources
// @Summary				get terraform state resources. This output is similar to "terraform state list"
// @Description.markdown	get_terraform_state_json_resources.md
// @Param					workspace_id	path	string	true	"workspace ID"
// @Param					state_id 		path	string	true	"state ID"
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
// @Success				200	{object}	interface{}
// @Router					/v1/runners/terraform-workspace/{workspace_id}/state-json/{state_id}/resources [get]
func (s *service) GetTerraformWorkspaceStateResources(ctx *gin.Context) {
	workspaceID := ctx.Param("workspace_id")
	stateID := ctx.Param("state_id")
	if workspaceID == "" || stateID == "" {
		ctx.Error(stderr.ErrInvalidRequest{
			Err: errors.New("workspace_id  or state_id was not set"),
		})

		return
	}

	state, err := s.GetTerraformStatesJSONById(ctx, workspaceID, ctx.Param("state_id"))
	if err != nil {
		ctx.Error(err)
		return
	}
	if state == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "terraform state not found"})
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

	ctx.JSON(http.StatusOK, listTerraformStateResources(statejson))
}

func listTerraformStateResources(state *tfjson.State) []string {
	if state.Values == nil || state.Values.RootModule == nil {
		return nil
	}

	var results []string

	var walkModule func(mod *tfjson.StateModule, prefix string)
	walkModule = func(mod *tfjson.StateModule, prefix string) {
		for _, res := range mod.Resources {
			fullAddr := res.Address
			if prefix != "" {
				fullAddr = prefix + "." + fullAddr
			}
			results = append(results, fullAddr)
		}
		for _, child := range mod.ChildModules {
			childPrefix := child.Address
			if prefix != "" {
				childPrefix = prefix + "." + childPrefix
			}
			walkModule(child, childPrefix)
		}
	}

	walkModule(state.Values.RootModule, "")
	sort.Strings(results)
	return results
}
