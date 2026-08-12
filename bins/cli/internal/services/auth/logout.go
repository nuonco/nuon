package auth

import (
	"context"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (a *Service) Logout(ctx context.Context, asJSON bool) error {
	// Clear the API token and URL from config
	a.cfg.Set("api_token", "")
	a.cfg.Set("api_url", "")

	// Write the updated config to file
	if err := a.cfg.WriteConfig(); err != nil {
		return err
	}

	if asJSON {
		ui.PrintJSON(map[string]string{"status": "logged_out"})
		return nil
	}

	ui.PrintLn("✅ Successfully logged out.")
	return nil
}
