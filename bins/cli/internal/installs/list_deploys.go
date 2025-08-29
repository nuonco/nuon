package installs

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListDeploys(ctx context.Context, installID string, offset, limit int, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	view := ui.NewListView()

	deploys, hasMore, err := s.listInstallDeploys(ctx, installID, offset, limit)
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
	view.RenderPaging(data, offset, limit, hasMore)
	return nil
}

func (s *Service) listInstallDeploys(ctx context.Context, installID string, offset, limit int) ([]*models.AppInstallDeploy, bool, error) {
	installs, hasMore, err := s.api.GetInstallDeploys(ctx, installID, &models.GetPaginatedQuery{
		Offset: offset,
		Limit:  limit,
	})
	if err != nil {
		return nil, false, err
	}
	return installs, hasMore, nil
}
