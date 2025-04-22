package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

// @ID						GetTerraformCurrentStateData
// @Summary				get current terraform
// @Description.markdown	get_terraform_current_state.md
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
// @Success				200	{object}	app.TerraformState
// @Router					/v1/terraform-backend [get]
func (s *service) GetTerraformCurrentStateData(ctx *gin.Context) {
	sid, err := s.GetStateID(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get state ID: %w", err))
		return
	}
	state, err := s.helpers.GetTerraformState(ctx, sid)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get terraform state: %w", err))
		return
	}

	if state == nil || state.Data == nil {
		ctx.JSON(http.StatusNoContent, "")
		return
	}

	tfStateData := &app.TerraformStateData{}
	err = json.Unmarshal(state.Data, tfStateData)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to unmarshal terraform state data: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, tfStateData)
}

// validateTerraformStateLock gets the current terraform state and validates the lock ID if a lock exists
func (s *service) validateTerraformStateLock(ctx *gin.Context, sid string, lockID string) (*app.TerraformState, error) {
	currentState, err := s.helpers.GetTerraformState(ctx, sid)
	if err != nil {
		return nil, fmt.Errorf("unable to get terraform state: %w", err)
	}

	if currentState != nil && string(currentState.Lock) != "" {
		currLockID, err := s.helpers.GetLockID(currentState.Lock)
		if err != nil {
			return nil, fmt.Errorf("unable to get lock id: %w", err)
		}

		if lockID != currLockID {
			return nil, fmt.Errorf("lock id mismatch: %s != %s", lockID, currLockID)
		}
	}

	return currentState, nil
}

func (s *service) GetStateID(ctx *gin.Context) (string, error) {
	workspaceID := ctx.Query("workspace_id")
	if workspaceID == "" {
		return "", fmt.Errorf("workspace_id is required")
	}

	workspace := &app.TerraformWorkspace{}
	res := s.db.WithContext(ctx).Model(&app.TerraformWorkspace{}).Where("id = ?", workspaceID).First(workspace)
	if res.Error != nil {
		return "", fmt.Errorf("unable to get workspace: %w", res.Error)
	}

	if workspace.ID == "" {
		return "", fmt.Errorf("workspace not found")
	}

	return s.helpers.GetStateID(workspaceID, sha256.New()), nil
}
