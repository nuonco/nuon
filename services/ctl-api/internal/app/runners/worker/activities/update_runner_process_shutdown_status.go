package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type UpdateRunnerProcessShutdownStatusRequest struct {
	ShutdownID        string                          `validate:"required"`
	Status            app.RunnerProcessShutdownStatus `validate:"required"`
	StatusDescription string
}

// @temporal-gen-v2 activity
// @by-field ShutdownID
func (a *Activities) UpdateRunnerProcessShutdownStatus(ctx context.Context, req UpdateRunnerProcessShutdownStatusRequest) (*app.RunnerProcessShutdown, error) {
	var current app.RunnerProcessShutdown
	if res := a.db.WithContext(ctx).First(&current, "id = ?", req.ShutdownID); res.Error != nil {
		return nil, fmt.Errorf("unable to get runner process shutdown: %w", res.Error)
	}

	newComposite := app.NewCompositeStatus(ctx, app.Status(req.Status))
	newComposite.StatusHumanDescription = req.StatusDescription
	newComposite.History = append([]app.CompositeStatus{current.CompositeStatus}, current.CompositeStatus.History...)
	newComposite.History[0].History = nil

	res := a.db.WithContext(ctx).
		Model(&app.RunnerProcessShutdown{ID: req.ShutdownID}).
		Updates(app.RunnerProcessShutdown{
			CompositeStatus: newComposite,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update runner process shutdown status: %w", res.Error)
	}

	var updated app.RunnerProcessShutdown
	if res := a.db.WithContext(ctx).First(&updated, "id = ?", req.ShutdownID); res.Error != nil {
		return nil, fmt.Errorf("unable to get updated runner process shutdown: %w", res.Error)
	}

	return &updated, nil
}
