package components

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Get(ctx context.Context, compID string) error {
	component, err := s.api.GetComponent(ctx, compID)
	if err != nil {
		return err
	}

	ui.Line(ctx, "%s - %s", component.ID, component.Name)
	return nil
}
