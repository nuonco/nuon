package auth

import (
	"context"

	"github.com/spf13/cobra"
)

func (a *Service) Logout(ctx context.Context) error {
	// Clear the API token and URL from config
	a.cfg.Set("api_token", "")
	a.cfg.Set("api_url", "")

	// Write the updated config to file
	if err := a.cfg.WriteConfig(); err != nil {
		return err
	}

	// Print success message
	cmd := &cobra.Command{}
	cmd.Printf("✅ Successfully logged out.\n")
	return nil
}
