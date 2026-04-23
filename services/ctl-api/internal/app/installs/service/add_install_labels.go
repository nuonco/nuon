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

type AddInstallLabelsRequest struct {
	Labels map[string]string `json:"labels" validate:"required"`
}

func (r *AddInstallLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						AddInstallLabels
// @Summary				add labels to an install
// @Description			Merge the provided labels into the install's existing labels. Existing keys are overwritten.
// @Param					install_id	path	string					true	"install ID"
// @Param					req			body	AddInstallLabelsRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.Install
// @Router					/v1/installs/{install_id}/labels [POST]
func (s *service) AddInstallLabels(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AddInstallLabelsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	var install app.Install
	if err := s.db.WithContext(ctx).First(&install, "id = ?", installID).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	install.Labels.Merge(labels.Labels(req.Labels))

	if err := s.db.WithContext(ctx).Model(&install).Select("labels").Updates(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update install labels: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}
