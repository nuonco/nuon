package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CreateDeploy(ctx context.Context, installID, buildID string, deployDeps, deployDependencies, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	req := &models.ServiceCreateInstallDeployRequest{
		BuildID:            buildID,
		DeployDependents:   deployDeps,
		DeployDependencies: deployDependencies,
	}

	aid, err := s.api.CreateInstallDeploy(ctx, installID, req)
	if err != nil {
		return ui.PrintError(err)
	}

	printActionResult(asJSON, fmt.Sprintf("successfully triggered deploy for install %s", aid.ID), actionResult{
		InstallID: installID,
		ID:        aid.ID,
		Status:    "deploy_triggered",
	})

	return nil
}
