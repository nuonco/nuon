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

	if len(builds) == 0 {
		ui.Line(ctx, "No builds found")
	} else {
		for _, build := range builds {
			ui.Line(ctx, "%s - %s", build.ID, build.Status)
		}
	}

	return nil
}
