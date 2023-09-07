package installs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) List(ctx context.Context, appID string) {
	view := ui.NewListView()

	installs := []*models.AppInstall{}
	err := error(nil)
	if appID == "" {
		installs, err = s.api.GetAllInstalls(ctx)
	} else {
		installs, err = s.api.GetAppInstalls(ctx, appID)
	}
	if err != nil {
		view.Error(err)
		return
	}

	data := [][]string{
		[]string{
			"id",
			"name",
			"status",
			"created at",
		},
	}
	for _, install := range installs {
		data = append(data, []string{
			install.ID,
			install.Name,
			install.Status,
			install.CreatedAt,
		})
	}
	view.Render(data)
}
