package apps

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, appID string, asJSON bool) {
	view := ui.NewGetView()

	app, err := s.api.GetApp(ctx, appID)
	if err != nil {
		view.Error(err)
	}

	if asJSON == true {
		j, _ := json.Marshal(app)
		fmt.Println(string(j))
	} else {
		view.Render([][]string{
			[]string{"id", app.ID},
			[]string{"name", app.Name},
			[]string{"status", app.Status},
			[]string{"created at", app.CreatedAt},
			[]string{"updated at", app.UpdatedAt},
			[]string{"created by", app.CreatedByID},
		})
	}
}
