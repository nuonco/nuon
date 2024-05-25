package migrations

import (
	"context"
	"fmt"
	"strings"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (a *Migrations) migration014AppInputDisplayName(ctx context.Context) error {
	var appInputs []*app.AppInput
	res := a.db.WithContext(ctx).
		Find(&appInputs)
	if res.Error != nil {
		return res.Error
	}

	for _, input := range appInputs {
		displayName := strings.ReplaceAll(input.Name, "_", " ")
		displayName = strings.ToTitle(displayName)

		res := a.db.WithContext(ctx).
			Model(&app.AppInput{
				ID: input.ID,
			}).
			Updates(app.AppInput{DisplayName: displayName})
		if res.Error != nil {
			return fmt.Errorf("unable to update install input to point to app input config: %w", res.Error)
		}
	}

	// hard delete any install inputs that were deleted
	var deletedAppInputs []app.AppInput
	res = a.db.WithContext(ctx).Unscoped().Find(&deletedAppInputs)
	if res.Error != nil {
		return fmt.Errorf("unable to get deleted app inputs: %w", res.Error)
	}

	if len(deletedAppInputs) < 1 {
		return nil
	}
	res = a.db.WithContext(ctx).Unscoped().Delete(&deletedAppInputs)
	if res.Error != nil {
		return fmt.Errorf("unable to delete app inputs: %w", res.Error)
	}

	return nil
}
