package helpers

import (
	"context"
	"maps"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// ApplyAppDefaultLabels reconciles the install's labels against the app's
// current default labels, using the install's snapshot to clean up defaults
// that were removed from the app config.
func (h *Helpers) ApplyAppDefaultLabels(ctx context.Context, installID string) error {
	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "app_id", "labels", "label_templates", "app_default_labels").
		First(&install, "id = ?", installID).Error; err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	var parentApp app.App
	if err := h.db.WithContext(ctx).
		Select("id", "default_labels").
		First(&parentApp, "id = ?", install.AppID).Error; err != nil {
		return errors.Wrap(err, "unable to get app")
	}

	newLabels, newTemplates, changed := labels.ApplyDefaults(
		install.Labels, install.LabelTemplates, install.AppDefaultLabels, parentApp.DefaultLabels)
	snapshotChanged := !maps.Equal(
		map[string]string(install.AppDefaultLabels), map[string]string(parentApp.DefaultLabels))
	if !changed && !snapshotChanged {
		return nil
	}

	updates := map[string]any{
		"labels":             newLabels,
		"label_templates":    newTemplates,
		"app_default_labels": parentApp.DefaultLabels,
	}
	if err := h.db.WithContext(ctx).Model(&app.Install{}).
		Where("id = ?", installID).
		Updates(updates).Error; err != nil {
		return errors.Wrap(err, "unable to persist default labels")
	}

	if len(newTemplates) > 0 {
		return h.RenderInstallLabels(ctx, installID)
	}

	return nil
}
