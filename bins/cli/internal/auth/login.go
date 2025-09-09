package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
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

	// Ask user about deployment type and hostname
	apiURL, err := a.selectAPIURL()
	if err != nil {
		return view.Error(fmt.Errorf("couldn't select API URL: %w", err))
	}

	// Set the API URL in the config
	cliCfg.Set("api_url", apiURL)
	cliCfg.APIURL = apiURL

	// Recreate the API client with the selected URL
	if err := a.updateAPIClient(apiURL, cliCfg); err != nil {
		return view.Error(fmt.Errorf("couldn't update API client: %w", err))
	}

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
	err = cliCfg.WriteConfig()
	if err != nil {
		return view.Error(err)
	}

	// get user info from ID token
	user := a.getUserInfo(tokens.IDToken)

	view.Render([][]string{
		{"Now logged in as", user.Name, user.Email},
	})

	return nil
}

// selectAPIURL prompts the user to select their deployment type and returns the appropriate API URL
func (a *Service) selectAPIURL() (string, error) {
	const defaultURL = "https://api.nuon.co"
	
	var deploymentType string
	prompt := &survey.Select{
		Message: "Which Nuon deployment are you using?",
		Options: []string{"Nuon Cloud", "BYOC Nuon"},
		Default: "Nuon Cloud",
	}
	
	if err := survey.AskOne(prompt, &deploymentType); err != nil {
		return "", fmt.Errorf("failed to get deployment type: %w", err)
	}
	
	if deploymentType == "Nuon Cloud" {
		return defaultURL, nil
	}
	
	// For BYOC Nuon, ask for custom hostname
	var customHostname string
	hostPrompt := &survey.Input{
		Message: "Enter your Nuon API hostname:",
		Help:    "Example: api.your-domain.com",
	}
	
	if err := survey.AskOne(hostPrompt, &customHostname, survey.WithValidator(survey.Required)); err != nil {
		return "", fmt.Errorf("failed to get custom hostname: %w", err)
	}
	
	customHostname = strings.TrimSpace(customHostname)
	if !strings.HasPrefix(customHostname, "http://") && !strings.HasPrefix(customHostname, "https://") {
		// Use http:// for localhost, https:// for everything else
		if strings.HasPrefix(customHostname, "localhost") || strings.HasPrefix(customHostname, "127.0.0.1") {
			customHostname = "http://" + customHostname
		} else {
			customHostname = "https://" + customHostname
		}
	}
	
	return customHostname, nil
}

// updateAPIClient recreates the API client with the new URL
func (a *Service) updateAPIClient(apiURL string, cliCfg *config.Config) error {
	// Create a new validator instance
	v := validator.New()
	
	// Create a new API client with the updated URL
	// Note: We don't have an API token yet since this is during login
	api, err := nuon.New(
		nuon.WithValidator(v),
		nuon.WithAuthToken(""), // Empty token during login
		nuon.WithOrgID(""),     // Empty org ID during login  
		nuon.WithURL(apiURL),
	)
	if err != nil {
		return fmt.Errorf("unable to create API client with URL %s: %w", apiURL, err)
	}

	// Update the service's API client
	a.api = api
	
	return nil
}
