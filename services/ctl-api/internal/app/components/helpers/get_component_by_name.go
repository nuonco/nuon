package helpers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
)

func (s *Helpers) GetComponentByName(ctx context.Context, appID, name string) (*app.Component, error) {
	cmp := app.Component{}
	res := s.db.WithContext(ctx).
		Where(app.Component{
			Name:  name,
			AppID: appID,
		}).
		First(&cmp)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component by name: %w", res.Error)
	}

	return &cmp, nil
}

func (s *Helpers) GetComponentIDs(ctx context.Context, appID string, comps []string) ([]string, error) {
	if len(comps) == 0 {
		return []string{}, nil
	}

	var components []app.Component
	res := s.db.WithContext(ctx).
		Select("id").
		Where("app_id = ?", appID).
		Where("name IN ? OR id IN ?", comps, comps).
		Find(&components)

	if res.Error != nil {
		return nil, errors.Wrap(res.Error, "unable to get components")
	}

	compIDs := make([]string, len(components))
	for i, comp := range components {
		compIDs[i] = comp.ID
	}

	return compIDs, nil
}
