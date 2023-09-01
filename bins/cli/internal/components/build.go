package components

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Build(ctx context.Context, compID string) error {
	build, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		return err
	}

	ui.Line(ctx, "Component build ID: %s", build.ID)
	return nil
}
