package orgs

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Current(ctx context.Context, asJSON bool) {
	view := ui.NewGetView()

	org, err := s.api.GetOrg(ctx)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON == true {
		j, _ := json.Marshal(org)
		fmt.Println(string(j))
	} else {
		view.Render([][]string{
			[]string{"id", org.ID},
			[]string{"name", org.Name},
			[]string{"status", org.StatusDescription},
			[]string{"created at", org.CreatedAt},
			[]string{"updated at", org.UpdatedAt},
			[]string{"created by", org.CreatedByID},
		})
	}
}
