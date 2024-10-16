package installs

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"

	"github.com/powertoolsdev/mono/bins/cli/internal/lookup"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

func (s *Service) DeployLogs(ctx context.Context, installID, deployID, installComponentID string, asJSON bool) error {
	installID, err := lookup.InstallID(ctx, s.api, installID)
	if err != nil {
		return ui.PrintError(err)
	}

	cfg, err := s.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	url := fmt.Sprintf("%s/%s/installs/%s/components/%s/deploys/%s", cfg.DashboardURL, s.cfg.OrgID, installID, installComponentID, deployID)
	var cmd *exec.Cmd

	// Determine the OS and set the command accordingly
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin": // macOS
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		// If the OS is not supported, print the URL
		ui.PrintLn("Use the following URL to view the logs")
		ui.PrintLn(url)
		return nil
	}

	return cmd.Start()
}
