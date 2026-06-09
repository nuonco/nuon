package runbooks

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (s *Service) Get(ctx context.Context, installID, runbookID string, asJSON bool) error {
	view := ui.NewGetView()

	rb, err := s.api.GetInstallRunbook(ctx, installID, runbookID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(rb)
		return nil
	}

	name := ""
	stepCount := 0
	if rb.Runbook != nil {
		name = rb.Runbook.Name
		if len(rb.Runbook.Configs) > 0 {
			stepCount = len(rb.Runbook.Configs[0].Steps)
		}
	}

	view.Render([][]string{
		{"id", rb.ID},
		{"name", name},
		{"runbook id", rb.RunbookID},
		{"status", rb.Status},
		{"steps", fmt.Sprintf("%d", stepCount)},
		{"created at", rb.CreatedAt},
		{"updated at", rb.UpdatedAt},
	})

	return nil
}
