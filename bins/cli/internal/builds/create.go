package builds

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) Create(ctx context.Context, compID string) {
	spinner := ui.NewSpinner()
	spinner.Start()
	basicText := ui.NewBasicText()

	spinner.UpdateText(basicText.Sprintfln("Starting build for component %s", compID))
	build, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		spinner.Fail(basicText.Sprintfln("build failed: %s", err))
		return
	}

	spinner.Success(basicText.Sprintfln("build completed: %s", build.ID))
	spinner.Stop()
}
