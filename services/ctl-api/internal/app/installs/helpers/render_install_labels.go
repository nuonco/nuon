package helpers

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/labels"
	"github.com/nuonco/nuon/pkg/render"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// RenderLabelTemplates renders label templates against current install state,
// returning only keys that rendered non-empty. Unresolvable templates (e.g. a
// component output not yet populated) are skipped with a warning, never an
// error. State is redacted so sensitive inputs can't leak into org-visible
// labels, and labels are stripped from the context to prevent self-reference.
func (h *Helpers) RenderLabelTemplates(ctx context.Context, installID string, templates labels.Labels) (labels.Labels, error) {
	rendered := make(labels.Labels, len(templates))
	if len(templates) == 0 {
		return rendered, nil
	}

	is, err := h.GetInstallState(ctx, installID, true, true)
	if err != nil {
		return nil, errors.Wrap(err, "unable to get install state")
	}

	is.Labels = nil
	data, err := is.AsMap()
	if err != nil {
		return nil, errors.Wrap(err, "unable to convert state to map")
	}

	for key, tmpl := range templates {
		out, err := render.RenderTextV2(tmpl, data)
		if err != nil || out == "" {
			cctx.GetLogger(ctx, h.l).Warn("unable to render install label template",
				zap.String("install_id", installID),
				zap.String("label_key", key),
				zap.Error(err))
			continue
		}
		rendered[key] = out
	}

	return rendered, nil
}

// RenderInstallLabels materializes the install's label templates into the
// labels column so label matching only ever sees literal values. Unresolvable
// keys keep their previous rendered value, or stay absent if they never
// rendered.
func (h *Helpers) RenderInstallLabels(ctx context.Context, installID string) error {
	var install app.Install
	if err := h.db.WithContext(ctx).
		Select("id", "labels", "label_templates").
		First(&install, "id = ?", installID).Error; err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	if len(install.LabelTemplates) == 0 {
		return nil
	}

	rendered, err := h.RenderLabelTemplates(ctx, installID, install.LabelTemplates)
	if err != nil {
		return err
	}

	merged := make(labels.Labels, len(install.Labels)+len(rendered))
	for k, v := range install.Labels {
		merged[k] = v
	}

	changed := false
	for key, val := range rendered {
		if merged[key] != val {
			merged[key] = val
			changed = true
		}
	}

	if !changed {
		return nil
	}

	if err := h.db.WithContext(ctx).Model(&app.Install{}).
		Where("id = ?", installID).
		Update("labels", merged).Error; err != nil {
		return errors.Wrap(err, "unable to persist rendered labels")
	}

	return nil
}
