package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) VCSConnections(ctx context.Context, asJSON bool) {
	view := ui.NewGetView()

	vcs, err := s.api.GetVCSConnections(ctx)
	if err != nil {
		view.Error(err)
		return
	}

	if asJSON {
		ui.PrintJSON(vcs)
		return
	}

	data := [][]string{
		{
			"GITHUB INSTALL ID",
		},
	}

	for _, v := range vcs {
		data = append(data, []string{
			*&v.GithubInstallID,
		})
	}

	view.Render(data)
}
