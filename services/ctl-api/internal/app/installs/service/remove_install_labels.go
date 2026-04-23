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

type RemoveInstallLabelsRequest struct {
	Keys []string `json:"keys" validate:"required"`
}

func (r *RemoveInstallLabelsRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(r); err != nil {
		return validatorPkg.FormatValidationError(err)
	}
	return nil
}

// @ID						RemoveInstallLabels
// @Summary				remove labels from an install
// @Description			Remove the specified label keys from the install.
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	RemoveInstallLabelsRequest	true	"Input"
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
// @Router					/v1/installs/{install_id}/labels [DELETE]
func (s *service) RemoveInstallLabels(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req RemoveInstallLabelsRequest
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

	install.Labels.RemoveKeys(req.Keys)

	if err := s.db.WithContext(ctx).Model(&install).Select("labels").Updates(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update install labels: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, install)
}
