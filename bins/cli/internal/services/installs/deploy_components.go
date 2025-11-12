package installs

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeployComponents(ctx context.Context, installID string, planOnly bool, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	err = s.api.DeployInstallComponents(ctx, installID, planOnly)
	if err != nil {
		fmt.Printf("deploy components err: %+s\n", err)
		return ui.PrintJSONError(err)
	}

	ui.PrintLn("successfully triggered deploy of all install components")
	return nil
}
