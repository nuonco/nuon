package releases

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) Create(ctx context.Context, compID, buildID, delay string, installsPerStep int64) {
	view := ui.NewCreateView("release")

	view.Start()
	release, err := s.api.CreateComponentRelease(ctx, compID, &models.ServiceCreateComponentReleaseRequest{
		BuildID: buildID,
		Strategy: &models.ServiceCreateComponentReleaseRequestStrategy{
			Delay:           delay,
			InstallsPerStep: installsPerStep,
		},
	})
	if err != nil {
		view.Fail(err)
		return
	}
	view.Success(release.ID)
}
