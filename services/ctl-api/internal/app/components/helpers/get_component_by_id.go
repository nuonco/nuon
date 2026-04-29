package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) GetComponentByID(ctx context.Context, componentID string) (*app.Component, error) {
	cmp := app.Component{}
	res := s.db.WithContext(ctx).
		Select("id, name").
		Where(app.Component{ID: componentID}).
		First(&cmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component by id: %w", res.Error)
	}

	return &cmp, nil
}
