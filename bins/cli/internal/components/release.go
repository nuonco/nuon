package components

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Release(ctx context.Context, compID, buildID string) error {
	release, err := s.api.CreateComponentRelease(ctx, compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			InstallsPerStep: 0,
		},
	})
	if err != nil {
		return err
	}

	ui.Line(ctx, "Component release ID: %s", release.ID)
	return nil
}
