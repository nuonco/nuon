package auth

import (
	"context"

	"github.com/spf13/cobra"
)

func (a *Service) Logout(ctx context.Context) error {
	a.cfg.Set("api_token", "")
	a.cfg.Set("api_url", "")

	if err := a.cfg.WriteConfig(); err != nil {
		return err
	}

	cmd := &cobra.Command{}
	cmd.Printf("✅ Successfully logged out.\n")
	return nil
}
