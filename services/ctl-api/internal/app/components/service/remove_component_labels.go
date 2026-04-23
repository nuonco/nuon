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

type RemoveComponentLabelsRequest struct {
	Keys []string `json:"keys" validate:"required"`
}

func (r *RemoveComponentLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						RemoveAppComponentLabels
// @Summary				remove labels from a component
// @Description			Remove the specified label keys from the component.
// @Param					app_id			path	string							true	"app ID"
// @Param					component_id	path	string							true	"component ID"
// @Param					req				body	RemoveComponentLabelsRequest	true	"Input"
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
// @Router					/v1/apps/{app_id}/components/{component_id}/labels [DELETE]
func (s *service) RemoveAppComponentLabels(ctx *gin.Context) {
	componentID := ctx.Param("component_id")

	var req RemoveComponentLabelsRequest
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

	component.Labels.RemoveKeys(req.Keys)

	if err := s.db.WithContext(ctx).Model(&component).Select("labels").Updates(&component).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update component labels: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, component)
}
