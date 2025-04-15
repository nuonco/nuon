package helpers

import (
	"context"
	"fmt"

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
