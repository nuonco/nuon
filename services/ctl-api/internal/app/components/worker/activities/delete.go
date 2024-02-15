package activities

import (
	"context"
	"errors"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"gorm.io/gorm"
)

type DeleteRequest struct {
	ComponentID string `validate:"required"`
}

func (a *Activities) Delete(ctx context.Context, req DeleteRequest) error {
	res := a.db.WithContext(ctx).Unscoped().Delete(&app.Component{
		ID: req.ComponentID,
	})
	if errors.Is(res.Error, gorm.ErrRecordNotFound) {
		return nil
	}

	if res.Error != nil {
		return fmt.Errorf("unable to delete component: %w", res.Error)
	}

	return nil
}
