package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	helpers "github.com/powertoolsdev/mono/bins/cli/internal"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListDeploys(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewGetView()

	deploys, err := s.listInstallDeploys(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(deploys)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"STATUS",
			"TYPE",
			"BUILD ID",
			"CREATED AT",
			"COMPONENT ID",
			"COMPONENT NAME",
			"COMPONENT CONFIG VERSION",
		},
	}
	for _, deploy := range deploys {
		data = append(data, []string{
			deploy.ID,
			deploy.Status,
			string(deploy.InstallDeployType),
			deploy.BuildID,
			deploy.CreatedAt,
			deploy.ComponentID,
			deploy.ComponentName,
			fmt.Sprintf("%d", deploy.ComponentConfigVersion),
		})
	}
	view.Render(data)
	return nil
}

func (s *Service) listInstallDeploys(ctx context.Context, installID string) ([]*models.AppInstallDeploy, error) {
	if !s.cfg.PaginationEnabled {
		cmps, _, err := s.api.GetInstallDeploys(ctx, installID, &models.GetInstallDeploysQuery{
			Offset:            0,
			Limit:             10,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, err
		}
		return cmps, nil
	}

	fetchFn := func(ctx context.Context, offset, limit int) ([]*models.AppInstallDeploy, bool, error) {
		installs, hasMore, err := s.api.GetInstallDeploys(ctx, installID, &models.GetInstallDeploysQuery{
			Offset:            offset,
			Limit:             limit,
			PaginationEnabled: s.cfg.PaginationEnabled,
		})
		if err != nil {
			return nil, false, err
		}
		return installs, hasMore, nil
	}

	return helpers.BatchFetch(ctx, 10, 50, fetchFn)
}
