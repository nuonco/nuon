package components

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Delete(ctx context.Context, compID string) error {
	_, err := s.api.DeleteComponent(ctx, compID)
	if err != nil {
		return err
	}

	ui.Line(ctx, "Component %s was deleted", compID)
	return nil
}
