package releases

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/powertoolsdev/mono/pkg/ui"
)

func (s *Service) Create(ctx context.Context, compID, buildID, delay string, installsPerStep int64) error {
	release, err := s.api.CreateComponentRelease(ctx, compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			Delay:           delay,
			InstallsPerStep: installsPerStep,
		},
	})
	if err != nil {
		return err
	}
	ui.Line(ctx, "%s - %s", release.ID, release.Status)

	return nil
}
