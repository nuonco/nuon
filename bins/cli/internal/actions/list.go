package actions

import (
	"context"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, appID string, asJSON bool) error {
	view := ui.NewListView()

	if appID == "" {
		s.printAppNotSetMsg()
		return nil
	}

	wfs, err := s.api.GetActionWorkflows(ctx, appID)

	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(wfs)
		return nil
	}

	data := [][]string{
		{
			"NAME",
			"ID",
		},
	}

	for _, wf := range wfs {
		data = append(data, []string{
			wf.Name,
			wf.ID,
		})
	}
	view.Render(data)
	return nil
}
