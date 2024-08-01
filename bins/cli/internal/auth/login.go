package auth

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

var (
	AuthDomain   string
	AuthClientID string
	AuthAudience string
)

func (a *Service) Login(ctx context.Context, cliCfg *config.Config) error {
	view := ui.NewGetView()

	cfg, err := a.api.GetCLIConfig(ctx)
	if err != nil {
		return view.Error(fmt.Errorf("couldn't get cli config: %w", err))
	}

	AuthAudience = cfg.AuthAudience
	AuthClientID = cfg.AuthClientID
	AuthDomain = cfg.AuthDomain

	// get device code
	deviceCode, err := a.getDeviceCode()
	if err != nil {
		return view.Error(fmt.Errorf("couldn't verify device code: %w", err))
	}

	tokens, err := a.getOAuthTokens(deviceCode)
	if err != nil {
		return view.Error(fmt.Errorf("couldn't get OAuth tokens: %w", err))
	}

	// add access token to config and write to the file
	cliCfg.Set("api_token", tokens.AccessToken)
	cliCfg.WriteConfig()

	// get user info from ID token
	user := a.getUserInfo(tokens.IDToken)

	view.Render([][]string{
		{"Now logged in as", user.Name, user.Email},
	})

	view.Render([][]string{
		{"Access token", tokens.AccessToken},
	})
	return nil
}
