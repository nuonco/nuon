package builds

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/pkg/api/client/models"
	"github.com/pterm/pterm"
)

func (s *Service) Create(ctx context.Context, compID string) error {
	pterm.DefaultSpinner.Start()

	pterm.DefaultSpinner.UpdateText(fmt.Sprintf("Starting build for component %s", compID))
	build, err := s.api.CreateComponentBuild(
		ctx,
		compID,
		&models.ServiceCreateComponentBuildRequest{
			UseLatest: true,
		},
	)
	if err != nil {
		pterm.DefaultSpinner.Fail(fmt.Sprintf("build failed: %s", err))
		return nil
	}

	pterm.DefaultSpinner.Success(fmt.Sprintf("build completed: %s", build.ID))

	pterm.DefaultSpinner.Stop()
	return nil
}
