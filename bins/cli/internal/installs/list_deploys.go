package installs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ListDeploys(ctx context.Context, installID string, asJSON bool) {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	view := ui.NewGetView()

	deploys, err := s.api.GetInstallDeploys(ctx, installID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(deploys)
		return
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
}
