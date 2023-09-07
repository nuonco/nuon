package components

import (
	"context"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
	"github.com/powertoolsdev/mono/pkg/api/client/models"
)

func (s *Service) List(ctx context.Context, appID string) {
	view := ui.NewListView()

	components := []*models.AppComponent{}
	err := error(nil)
	if appID != "" {
		components, err = s.api.GetAppComponents(ctx, appID)
	} else {
		components, err = s.api.GetAllComponents(ctx)
	}
	if err != nil {
		view.Error(err)
		return
	}

	data := [][]string{
		[]string{
			"id",
			"name",
			"created at",
			"updated at",
			"created by",
			"config versions",
		},
	}
	for _, component := range components {
		data = append(data, []string{
			component.ID,
			component.Name,
			component.CreatedAt,
			component.UpdatedAt,
			component.CreatedByID,
			strconv.Itoa(int(component.ConfigVersions)),
		})
	}
	view.Render(data)
}
