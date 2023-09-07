package components

import (
	"context"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, compID string) {
	view := ui.NewGetView()

	component, err := s.api.GetComponent(ctx, compID)
	if err != nil {
		view.Error(err)
		return
	}

	view.Render([][]string{
		[]string{"id", component.ID},
		[]string{"name", component.Name},
		[]string{"created at", component.CreatedAt},
		[]string{"updated at", component.UpdatedAt},
		[]string{"created by", component.CreatedByID},
		[]string{"app id ", component.AppID},
		[]string{"config versions", strconv.Itoa(int(component.ConfigVersions))},
	})
}
