package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/labeladded"
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
// @Description			Merge the provided labels into the install's existing labels. Existing keys are overwritten. A value using the .nuon interpolation syntax becomes a dynamic label: the template is stored and its rendered value is re-materialized whenever install state changes. Keys managed by the app config's default_labels cannot be changed here.
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

	// Default labels are set in the app config; a write that echoes the current
	// rendered value or the default itself is a harmless round-trip, anything
	// else is rejected.
	for key, val := range req.Labels {
		def, isDefault := install.AppDefaultLabels[key]
		if !isDefault {
			continue
		}
		if val == install.Labels[key] || val == def {
			delete(req.Labels, key)
			continue
		}
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("label %q is a default label; edit default_labels in the app config and sync", key),
			Description: fmt.Sprintf("label %q is a default label; edit default_labels in the app config and sync", key),
		})
		return
	}

	static, templated := labels.Labels(req.Labels).SplitTemplated()
	for key, tmpl := range templated {
		if err := render.ValidateTextTemplate(tmpl); err != nil {
			ctx.Error(stderr.ErrUser{
				Err:         fmt.Errorf("label %q uses the interpolation syntax but is not a valid template: %w", key, err),
				Description: fmt.Sprintf("label %q uses the interpolation syntax but is not a valid template", key),
			})
			return
		}
	}

	// A static write echoing a template-managed key's rendered value is a
	// round-trip from a client that only sees rendered labels — keep the
	// template. A differing value converts the key back to static.
	for key, val := range static {
		if _, managed := install.LabelTemplates[key]; managed && install.Labels[key] == val {
			delete(static, key)
		}
	}

	newTemplates := make(labels.Labels, len(install.LabelTemplates)+len(templated))
	for k, v := range install.LabelTemplates {
		newTemplates[k] = v
	}
	newTemplates.RemoveKeys(mapKeys(static))
	newTemplates.Merge(templated)

	merged := make(labels.Labels, len(install.Labels)+len(req.Labels))
	for k, v := range install.Labels {
		merged[k] = v
	}
	merged.Merge(static)

	if len(templated) > 0 {
		rendered, err := s.helpers.RenderLabelTemplates(ctx, installID, templated)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to render label templates: %w", err))
			return
		}
		merged.Merge(rendered)
	}

	if err := s.appsHelpers.ValidateInstallBranchExclusivity(ctx, &install, merged); err != nil {
		ctx.Error(err)
		return
	}

	changedLabelNames := make([]string, 0, len(req.Labels))
	for key := range req.Labels {
		if oldValue, ok := install.Labels[key]; !ok || oldValue != merged[key] {
			changedLabelNames = append(changedLabelNames, key)
		}
	}

	install.Labels = merged
	install.LabelTemplates = newTemplates

	if err := s.db.WithContext(ctx).Model(&install).Select("labels", "label_templates").Updates(&install).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to update install labels: %w", err))
		return
	}

	if len(changedLabelNames) > 0 {
		queueID, err := s.getInstallSignalsQueueID(ctx, install.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		for _, labelName := range changedLabelNames {
			if err := s.enqueueInstallSignal(ctx, queueID, &labeladded.Signal{
				InstallID: install.ID,
				LabelName: labelName,
			}, "", ""); err != nil {
				ctx.Error(fmt.Errorf("unable to enqueue label-added signal: %w", err))
				return
			}
		}
	}

	ctx.JSON(http.StatusOK, install)
}

func mapKeys(m labels.Labels) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
