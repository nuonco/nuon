package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type RemoveActionLabelsRequest struct {
	Keys []string `json:"keys" validate:"required"`
}

func (r *RemoveActionLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						RemoveAppActionLabels
// @Summary				remove labels from an action
// @Description			Remove the specified label keys from the action.
// @Param					app_id		path	string						true	"app ID"
// @Param					action_id	path	string						true	"action ID"
// @Param					req			body	RemoveActionLabelsRequest	true	"Input"
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.ActionWorkflow
// @Router					/v1/apps/{app_id}/actions/{action_id}/labels [DELETE]
func (s *service) RemoveAppActionLabels(ctx *gin.Context) {
	actionID := ctx.Param("action_id")

	var req RemoveActionLabelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	var action app.ActionWorkflow
	if err := s.db.WithContext(ctx).First(&action, "id = ?", actionID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get action %s: %w", actionID, err))
		return
	}

	action.Labels.RemoveKeys(req.Keys)

	if err := s.db.WithContext(ctx).Model(&action).Select("labels").Updates(&action).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update action labels: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, action)
}
