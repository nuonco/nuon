package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type DeleteRequest struct {
	AppID string `validate:"required"`
}

func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	res := a.db.WithContext(ctx).Delete(&app.App{
		ID: req.AppID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete app: %w", res.Error)
	}

	return nil
}
