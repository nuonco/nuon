package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Current(ctx context.Context, asJSON bool) {
	view := ui.NewGetView()

	org, err := s.api.GetOrg(ctx)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(org)
		return
	}

	view.Render([][]string{
		{"id", org.ID},
		{"name", org.Name},
		{"status", org.StatusDescription},
		{"created at", org.CreatedAt},
		{"updated at", org.UpdatedAt},
		{"created by", org.CreatedByID},
	})
}
