package helpers

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) GetAppComponents(ctx context.Context, appID string) ([]app.Component, error) {
	var comps []app.Component
	if res := s.db.WithContext(ctx).
		Where("app_id = ?", appID).Find(&comps); res.Error != nil {
		return nil, fmt.Errorf("unable to get components: %w", res.Error)
	}

	return comps, nil
}
