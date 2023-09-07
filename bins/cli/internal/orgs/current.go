package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Current(ctx context.Context) {
	view := ui.NewGetView()

	org, err := s.api.GetOrg(ctx)
	if err != nil {
		view.Error(err)
		return
	}
	view.Render([][]string{
		[]string{"id", org.ID},
		[]string{"name", org.Name},
		[]string{"status", org.StatusDescription},
		[]string{"created at", org.CreatedAt},
		[]string{"updated at", org.UpdatedAt},
		[]string{"created by", org.CreatedByID},
	})
}
