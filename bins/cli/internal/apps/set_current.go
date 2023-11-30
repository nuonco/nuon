package apps

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) SetCurrent(ctx context.Context, appID string, cfg *config.Config) {
	appID, err := lookup.AppID(ctx, s.api, appID)
	if err != nil {
		ui.PrintError(err)
		return
	}

	cfg.Set("app_id", appID)
	cfg.WriteConfig()
	fmt.Printf("%s is now the current app\n", appID)
}
