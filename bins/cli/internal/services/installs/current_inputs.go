package installs

import (
	"context"
	"fmt"
	"sort"

	"charm.land/lipgloss/v2"
	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (s *Service) CurrentInputs(ctx context.Context, installID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}
	view := ui.NewGetView()

	inputs, err := s.listInstallInputs(ctx, installID)
	if err != nil {
		return view.Error(err)
	}

	if asJSON {
		ui.PrintJSON(inputs)
		return nil
	}

	for _, inp := range inputs {
		data := [][]string{}
		keys := make([]string, 0, len(inp.RedactedValues))
		for k := range inp.RedactedValues {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			data = append(data, []string{k, inp.RedactedValues[k]})
		}
		fmt.Println("")
		highlight := lipgloss.NewStyle().Foreground(styles.AccentColor).Bold(true)
		fmt.Println("inputs ID: " + highlight.Render(inp.ID))
		fmt.Println("modified at: " + highlight.Render(inp.CreatedAt))
		view.Render(data)
	}
	return nil
}

func (s *Service) listInstallInputs(ctx context.Context, installID string) ([]*models.AppInstallInputs, error) {
	inputs, _, err := s.api.GetInstallInputs(ctx, installID, &models.GetPaginatedQuery{
		Offset: 0,
		Limit:  1,
	})
	if err != nil {
		return nil, err
	}
	return inputs, nil
}
