package apps

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) ExportTerraform(ctx context.Context, appID string, asJSON bool) {
	view := ui.NewListView()

	cfgs, err := s.api.GetAppLatestConfig(ctx, appID)
	if err != nil {
		view.Error(err)
		return
	}

	fmt.Print(cfgs.GeneratedTerraform)
}
