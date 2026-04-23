package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	validatorPkg "github.com/nuonco/nuon/services/ctl-api/internal/pkg/validator"
)

type AddComponentLabelsRequest struct {
	Labels map[string]string `json:"labels" validate:"required"`
}

func (r *AddComponentLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AddAppComponentLabels
// @Summary				add labels to a component
// @Description			Merge the provided labels into the component's existing labels. Existing keys are overwritten.
// @Param					app_id			path	string						true	"app ID"
// @Param					component_id	path	string						true	"component ID"
// @Param					req				body	AddComponentLabelsRequest	true	"Input"
// @Tags					components
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Component
// @Router					/v1/apps/{app_id}/components/{component_id}/labels [POST]
func (s *service) AddAppComponentLabels(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	var req AddComponentLabelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	var component app.Component
	if err := s.db.WithContext(ctx).First(&component, "id = ?", componentID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get component %s: %w", componentID, err))
		return
	}

	component.Labels.Merge(labels.Labels(req.Labels))

	if err := s.db.WithContext(ctx).Model(&component).Select("labels").Updates(&component).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update component labels: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}
