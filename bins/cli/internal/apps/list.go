package apps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, asJSON bool) {
	view := ui.NewListView()

	apps, err := s.api.GetApps(ctx)
	if err != nil {
		view.Error(err)
	}

	if asJSON == true {
		j, _ := json.Marshal(apps)
		fmt.Println(string(j))
	} else {
		data := [][]string{
			[]string{
				"id",
				"name",
				"status",
			},
		}
		for _, app := range apps {
			data = append(data, []string{
				app.ID,
				app.Name,
				app.Status,
			})
		}
		view.Render(data)
	}
}
