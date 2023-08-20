package activities

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

type DeleteRequest struct {
	ComponentID string `validate:"required"`
}

func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	res := a.db.WithContext(ctx).Unscoped().Delete(&app.Component{
		ID: req.ComponentID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete component: %w", res.Error)
	}

	// NOTE(jm): since this inevitably an async operation, we do not error if the app was not found when deleting,
	// as the parent org could have deleted this first.
	//
	// Eventually, we would want the parent org to ensure all child component workflows are closed + deleted, but
	// for now this is not guaranteed.
	return nil
}
