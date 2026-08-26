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
// @Description			Remove the specified label keys from the install. Removing a dynamic label's key also removes its template. Keys managed by the app config's default_labels cannot be removed here.
// @Param					install_id	path	string						true	"install ID"
// @Param					req			body	RemoveInstallLabelsRequest	true	"Input"
// @Tags					installs
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
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

	for _, key := range req.Keys {
		if _, isDefault := install.AppDefaultLabels[key]; isDefault {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("label %q is a default label; remove it from default_labels in the app config and sync", key),
				Description: fmt.Sprintf("label %q is a default label; remove it from default_labels in the app config and sync", key),
			})
			return
		}
	}

	install.Labels.RemoveKeys(req.Keys)
	install.LabelTemplates.RemoveKeys(req.Keys)

	if err := s.db.WithContext(ctx).Model(&install).Select("labels", "label_templates").Updates(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update install labels: %w", err))
		return
	}

	matches, _ := s.appsHelpers.FindBranchesMatchingLabels(ctx, install.AppID, install.Labels)
	if len(matches) == 0 {
		s.appsHelpers.DeactivateInstallBranchConnections(ctx, install.ID)
	} else if len(matches) == 1 {
		s.appsHelpers.SyncInstallBranchConnection(ctx, &install, matches[0].Branch.ID)
	}

	ctx.JSON(http.StatusOK, install)
}
