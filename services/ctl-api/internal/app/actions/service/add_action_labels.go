package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	applookup "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/lookup"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AddActionLabelsRequest struct {
	Labels map[string]string `json:"labels" validate:"required"`
}

func (r *AddActionLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AddAppActionLabels
// @Summary				add labels to an action
// @Description			Merge the provided labels into the action's existing labels. Existing keys are overwritten.
// @Param					app_id		path	string					true	"app ID"
// @Param					action_id	path	string					true	"action ID"
// @Param					req			body	AddActionLabelsRequest	true	"Input"
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
// @Router					/v1/apps/{app_id}/actions/{action_id}/labels [POST]
func (s *service) AddAppActionLabels(ctx *gin.Context) {
	actionID := ctx.Param("action_id")

	var req AddActionLabelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	action, err := s.findAppLabelAction(ctx, actionID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get action %s: %w", actionID, err))
		return
	}

	action.Labels.Merge(labels.Labels(req.Labels))

	if err := s.db.WithContext(ctx).Model(action).Select("labels").Updates(action).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update action labels: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, action)
}

func (s *service) findAppLabelAction(ctx *gin.Context, actionID string) (*app.ActionWorkflow, error) {
	orgID, err := cctx.OrgIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	currentApp, err := applookup.ByRef(ctx, s.db, orgID, ctx.Param("app_id"))
	if err != nil {
		return nil, err
	}

	var action app.ActionWorkflow
	if err := s.db.WithContext(ctx).
		Where(&app.ActionWorkflow{ID: actionID, AppID: currentApp.ID, OrgID: orgID}, "ID", "AppID", "OrgID").
		First(&action).Error; err != nil {
		return nil, err
	}
	return &action, nil
}
