package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) GetComponents(ctx context.Context, compIDs []string) ([]app.Component, error) {
	var comps []app.Component
	if res := s.db.WithContext(ctx).
		Where("id IN ?", compIDs).Find(&comps); res.Error != nil {
		return nil, fmt.Errorf("unable to get components: %w", res.Error)
	}

	return comps, nil
}
