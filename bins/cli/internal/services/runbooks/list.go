package runbooks

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) List(ctx context.Context, installID string, asJSON bool) error {
	view := ui.NewListView()

	runbooks, err := s.api.GetInstallRunbooks(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(runbooks)
		return nil
	}

	data := [][]string{
		{
			"NAME",
			"ID",
			"STATUS",
			"STEPS",
		},
	}

	for _, rb := range runbooks {
		name := ""
		stepCount := 0
		if rb.Runbook != nil {
			name = rb.Runbook.Name
			if len(rb.Runbook.Configs) > 0 {
				stepCount = len(rb.Runbook.Configs[0].Steps)
			}
		}
		data = append(data, []string{
			name,
			rb.RunbookID,
			rb.Status,
			fmt.Sprintf("%d", stepCount),
		})
	}
	view.Render(data)
	return nil
}
