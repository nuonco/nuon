package orgs

import (
	"context"
	"strconv"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) Current(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	org, err := s.api.GetOrg(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(org)
		return nil
	}

	view.Render([][]string{
		{"id", org.ID},
		{"name", org.Name},
		{"status", org.StatusDescription},
		{"health-check status", string(org.LatestHealthCheck.Status)},
		{"latest health-check", org.LatestHealthCheck.CreatedAt},
		{"sandbox mode", strconv.FormatBool(org.SandboxMode)},
		{"created at", org.CreatedAt},
		{"updated at", org.UpdatedAt},
		{"created by", org.CreatedByID},
	})
	return nil
}
