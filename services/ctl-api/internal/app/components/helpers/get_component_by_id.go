package helpers

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) GetComponentByID(ctx context.Context, componentID string) (*app.Component, error) {
	cmp := app.Component{}
	// created_by_id and org_id are required by callers that set the account/org
	// context from the returned component before writing records (e.g. the
	// CreateComponentBuild activity, which otherwise inserts a build with a null
	// created_by_id and violates the NOT NULL constraint).
	res := s.db.WithContext(ctx).
		Select("id, name, created_by_id, org_id").
		Where(app.Component{ID: componentID}).
		First(&cmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component by id: %w", res.Error)
	}

	return &cmp, nil
}
