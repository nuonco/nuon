package builds

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, compID string, asJSON bool) {
	view := ui.NewListView()

	builds, err := s.api.GetComponentBuilds(ctx, compID)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON == true {
		j, _ := json.Marshal(builds)
		fmt.Println(string(j))
	} else {
		data := [][]string{
			[]string{
				"id",
				"status",
				"component id",
				"git ref",
			},
		}
		for _, build := range builds {
			data = append(data, []string{
				build.ID,
				build.Status,
				compID,
				build.GitRef,
			})
		}
		view.Render(data)
	}
}
