package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) UpdateComponentType(ctx context.Context, cmpID string, cmpType app.ComponentType) error {
	res := s.db.WithContext(ctx).Model(&app.Component{}).
		Where("id = ?", cmpID).
		Updates(app.Component{
			Type: cmpType,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update component type: %w", res.Error)
	}

	return nil
}
