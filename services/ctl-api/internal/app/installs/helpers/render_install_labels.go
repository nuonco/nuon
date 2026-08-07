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

// RenderInstallLabels renders the install's label templates against current
// install state and materializes the results into the labels column, so label
// matching (SQL containment, subscription dispatch, branch selectors) only
// ever sees literal values. Called after state (re)generation and after a
// config sync writes new templates.
//
// A template that cannot be resolved yet — e.g. it references a component
// output that has not populated — keeps the key's previous rendered value, or
// leaves the key absent if it never rendered. Render failures never propagate:
// labels must not fail state generation or config sync.
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

	// Redacted state so a template referencing a sensitive input can never
	// leak its value into a label, which is visible org-wide and in events.
	is, err := h.GetInstallState(ctx, installID, true, true)
	if err != nil {
		return errors.Wrap(err, "unable to get install state")
	}

	// Label templates must not reference labels themselves.
	is.Labels = nil
	data, err := is.AsMap()
	if err != nil {
		return errors.Wrap(err, "unable to convert state to map")
	}

	merged := make(labels.Labels, len(install.Labels)+len(install.LabelTemplates))
	for k, v := range install.Labels {
		merged[k] = v
	}

	changed := false
	for key, tmpl := range install.LabelTemplates {
		rendered, err := render.RenderTextV2(tmpl, data)
		if err != nil || rendered == "" {
			cctx.GetLogger(ctx, h.l).Warn("unable to render install label template",
				zap.String("install_id", installID),
				zap.String("label_key", key),
				zap.Error(err))
			continue
		}
		if merged[key] != rendered {
			merged[key] = rendered
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
