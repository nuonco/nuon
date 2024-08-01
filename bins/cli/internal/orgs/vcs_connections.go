package orgs

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) VCSConnections(ctx context.Context, asJSON bool) error {
	view := ui.NewGetView()

	vcs, err := s.api.GetVCSConnections(ctx)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(vcs)
		return nil
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
	return nil
}
