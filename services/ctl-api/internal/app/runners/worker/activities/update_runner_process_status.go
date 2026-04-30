package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateRunnerProcessStatusRequest struct {
	ProcessID         string                  `validate:"required"`
	Status            app.RunnerProcessStatus `validate:"required"`
	StatusDescription string
	Metadata          map[string]any
}

// @temporal-gen-v2 activity
// @by-field ProcessID
func (a *Activities) UpdateRunnerProcessStatus(ctx context.Context, req UpdateRunnerProcessStatusRequest) (*app.RunnerProcess, error) {
	var current app.RunnerProcess
	if res := a.db.WithContext(ctx).First(&current, "id = ?", req.ProcessID); res.Error != nil {
		return nil, fmt.Errorf("unable to get runner process: %w", res.Error)
	}

	newComposite := app.NewCompositeStatus(ctx, app.Status(req.Status))
	newComposite.StatusHumanDescription = req.StatusDescription
	for k, v := range current.CompositeStatus.Metadata {
		newComposite.Metadata[k] = v
	}
	for k, v := range req.Metadata {
		newComposite.Metadata[k] = v
	}
	newComposite.History = append([]app.CompositeStatus{current.CompositeStatus}, current.CompositeStatus.History...)
	newComposite.History[0].History = nil

	res := a.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: req.ProcessID}).
		Updates(app.RunnerProcess{
			CompositeStatus: newComposite,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update runner process status: %w", res.Error)
	}

	var updated app.RunnerProcess
	if res := a.db.WithContext(ctx).Preload("Shutdowns").First(&updated, "id = ?", req.ProcessID); res.Error != nil {
		return nil, fmt.Errorf("unable to get updated runner process: %w", res.Error)
	}

	return &updated, nil
}
