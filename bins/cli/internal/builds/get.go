package builds

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Get(ctx context.Context, compID, buildID string) error {
	build, err := s.api.GetComponentBuild(ctx, compID, buildID)
	if err != nil {
		return err
	}

	ui.Line(ctx, "%s - %s", build.ID, build.Status)
	return nil
}
