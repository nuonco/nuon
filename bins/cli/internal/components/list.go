package components

import (
	"context"
	"strconv"

	"github.com/nuonco/nuon-go/models"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appNameOrID string, asJSON bool) error {
	view := ui.NewListView()

	var (
		components []*models.AppComponent
		err        error
	)
	if appNameOrID != "" {
		appID, err := lookup.AppID(ctx, s.api, appNameOrID)
		if err != nil {
			return view.Error(err)
		}
		components, err = s.api.GetAppComponents(ctx, appID)
	} else {
		components, err = s.api.GetAllComponents(ctx)
	}
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(components)
		return nil
	}

	data := [][]string{
		{
			"ID",
			"NAME",
			"CREATED AT",
			"UPDATED AT",
			"CREATED BY",
			"CONFIG VERSIONS",
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
	return nil
}
