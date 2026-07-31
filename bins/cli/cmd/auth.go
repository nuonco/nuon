package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/bins/cli/internal/oidctoken"
	"github.com/nuonco/nuon/bins/cli/internal/services/auth"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

func (c *cli) authCmd() *cobra.Command {
	authCmd := &cobra.Command{
		Use:               "auth",
		Short:             "Login and logout commands",
		PersistentPreRunE: c.persistentPreRunE,
		GroupID:           CoreGroup.ID,
	}

	// Add login subcommand
	loginCmd := &cobra.Command{
		Use:         "login",
		Short:       "Login to Nuon",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := auth.New(c.apiClient, c.cfg)
			return svc.Login(cmd.Context())
		}),
	}

	// Add logout subcommand
	logoutCmd := &cobra.Command{
		Use:         "logout",
		Short:       "Logout from Nuon",
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			svc := auth.New(c.apiClient, c.cfg)
			return svc.Logout(cmd.Context())
		}),
	}

	var (
		oidcToken string
		audience  string
		orgID     string
	)
	exchangeTokenCmd := &cobra.Command{
		Use:   "exchange-token",
		Short: "Exchange an OIDC token for a Nuon API token",
		Long: `Exchange an OIDC ID token for a short-lived Nuon API token using your org's OIDC trust policies.

The OIDC token is resolved from --oidc-token, NUON_OIDC_TOKEN, NUON_OIDC_TOKEN_FILE, or the
GitHub Actions ID token endpoint (requires "permissions: id-token: write" in the workflow).

With the default table output, only the token is printed so it can be captured directly:

  export NUON_API_TOKEN=$(nuon auth exchange-token)`,
		Annotations: skipAuthAnnotation(),
		Run: c.wrapCmd(func(cmd *cobra.Command, _ []string) error {
			token := oidcToken
			if token == "" {
				detected, source, ok, err := oidctoken.Detect(cmd.Context(), oidctoken.Audience(audience))
				if err != nil {
					return ui.PrintError(&ui.CLIUserError{Msg: fmt.Sprintf("unable to get OIDC token from %s: %v", source, err)})
				}
				if !ok {
					return ui.PrintError(&ui.CLIUserError{Msg: "no OIDC token found: pass --oidc-token, set NUON_OIDC_TOKEN or NUON_OIDC_TOKEN_FILE, or run in GitHub Actions with `permissions: id-token: write`"})
				}
				token = detected
			}

			svc := auth.New(c.apiClient, c.cfg)
			resp, err := svc.ExchangeOIDCToken(cmd.Context(), token, orgID)
			if err != nil {
				if _, ok := nuon.ToAPIError(err); ok {
					return ui.PrintError(err)
				}
				return ui.PrintError(&ui.CLIUserError{Msg: err.Error()})
			}

			if PrintJSON {
				ui.PrintJSON(resp)
				return nil
			}

			fmt.Println(resp.Token)
			return nil
		}),
	}
	exchangeTokenCmd.Flags().StringVar(&oidcToken, "oidc-token", "", "The OIDC ID token to exchange (defaults to ambient detection)")
	exchangeTokenCmd.Flags().StringVar(&audience, "audience", "", "The audience to request for ambient OIDC tokens (or NUON_OIDC_AUDIENCE)")
	exchangeTokenCmd.Flags().StringVar(&orgID, "org-id", "", "The org to exchange the token with (defaults to the selected org)")

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(exchangeTokenCmd)

	return authCmd
}
