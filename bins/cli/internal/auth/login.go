package auth

import (
	"context"
	"fmt"

	"github.com/powertoolsdev/mono/bins/cli/internal/ui"
)

var (
	AuthDomain   string
	AuthClientID string
	AuthAudience string
)

func (a *Service) Login(ctx context.Context) {
	view := ui.NewGetView()

	cfg, err := a.api.GetCLIConfig(ctx)
	if err != nil {
		view.Error(fmt.Errorf("couldn't get cli config: %w", err))
		return
	}

	AuthAudience = cfg.AuthAudience
	AuthClientID = cfg.AuthClientID
	AuthDomain = cfg.AuthDomain

	// get device code
	deviceCode, err := a.getDeviceCode()
	if err != nil {
		view.Error(fmt.Errorf("couldn't verify device code: %w", err))
	}

	tokens, err := a.getOAuthTokens(deviceCode)
	if err != nil {
		view.Error(fmt.Errorf("couldn't get OAuth tokens: %w", err))
	}

	// write access token in config
	// TODO

	// get user info from ID token
	user := a.getUserInfo(tokens.IDToken)

	view.Render([][]string{
		{"Now logged in as", user.Name, user.Email},
	})

	view.Render([][]string{
		{"Access token", tokens.AccessToken},
	})
}
