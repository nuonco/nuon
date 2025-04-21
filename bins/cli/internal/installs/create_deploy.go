package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) CreateDeploy(ctx context.Context, installID, buildID string, deployDeps, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	req := &models.ServiceCreateInstallDeployRequest{
		BuildID:          buildID,
		DeployDependents: deployDeps,
	}

	aid, err := s.api.CreateInstallDeploy(ctx, installID, req)
	if err != nil {
		return ui.PrintError(err)
	}

	ui.PrintLn(fmt.Sprintf("successfully triggered deploy for install %s", aid.ID))

	return nil
}
