package builds

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) List(ctx context.Context, compID string) error {
	builds, err := s.api.GetComponentBuilds(ctx, compID)
	if err != nil {
		return err
	}

	for _, build := range builds {
		ui.Line(ctx, "%s - %s", build.ID, build.Status)
	}

	return nil
}
