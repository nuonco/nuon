package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/cli/styles"
)

var (
	AuthDomain   string
	AuthClientID string
	AuthAudience string
)

// LoginResult contains the result of an authentication flow
type LoginResult struct {
	AccessToken string
	DisplayName string
}

func (a *Service) Login(ctx context.Context) error {
	// Ask user about deployment type and hostname
	apiURL, err := a.selectAPIURL()
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't select API URL: %w", err))
	}

	// Set the API URL in the config
	a.cfg.Set("api_url", apiURL)
	a.cfg.APIURL = apiURL

	// Recreate the API client with the selected URL
	if err := a.updateAPIClient(apiURL, a.cfg); err != nil {
		return ui.PrintError(fmt.Errorf("couldn't update API client: %w", err))
	}

	cfg, err := a.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	// Determine which auth flow to use based on NuonAuthEnabled
	var result *LoginResult
	if cfg.NuonAuthEnabled {
		result, err = a.loginWithNuonAuth(ctx, cfg)
	} else {
		result, err = a.loginWithAuth0(ctx, cfg)
	}
	if err != nil {
		return ui.PrintError(err)
	}

	// Save access token to config
	a.cfg.Set("api_token", result.AccessToken)
	if err := a.cfg.WriteConfig(); err != nil {
		return ui.PrintError(err)
	}

	ui.PrintLn(fmt.Sprintf("Logged in as %s", result.DisplayName))

	// Update apiClient with newly-fetched token so we can list orgs
	api, err := nuon.New(
		nuon.WithValidator(validator.New()),
		nuon.WithAuthToken(result.AccessToken),
		nuon.WithURL(a.cfg.APIURL),
	)
	if err != nil {
		return ui.PrintError(fmt.Errorf("unable to init API client: %w", err))
	}
	a.api = api

	// If user only has a single org, select it
	orgs, _, err := a.api.GetOrgs(ctx, &models.GetPaginatedQuery{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return ui.PrintError(err)
	}

	switch len(orgs) {
	case 0:
		// prompt user to create an org
		ui.PrintLn("You are not a member of any orgs. You must create an org, or request an invite to one to continue.")

	case 1:
		org := orgs[0]
		a.cfg.Set("org_id", org.ID)
		err = a.cfg.WriteConfig()
		if err != nil {
			return ui.PrintError(err)
		}
		ui.PrintLn(fmt.Sprintf("Using org %s", org.Name))

	default:
		ui.PrintLn("You are a member of multiple orgs. Select one to continue.")
	}

	return nil
}

// loginWithAuth0 performs the Auth0 device code flow
func (a *Service) loginWithAuth0(ctx context.Context, cfg *models.ServiceCLIConfig) (*LoginResult, error) {
	AuthAudience = cfg.AuthAudience
	AuthClientID = cfg.AuthClientID
	AuthDomain = cfg.AuthDomain

	// Get device code
	deviceCode, err := a.getAuth0DeviceCode()
	if err != nil {
		return nil, fmt.Errorf("couldn't get device code: %w", err)
	}

	// Poll for tokens
	tokens, err := a.getOAuthTokens(deviceCode)
	if err != nil {
		return nil, fmt.Errorf("couldn't get OAuth tokens: %w", err)
	}

	// Get user info from ID token
	user := a.getUserInfo(tokens.IDToken)

	return &LoginResult{
		AccessToken: tokens.AccessToken,
		DisplayName: user.Name,
	}, nil
}

// selectAPIURL checks for a configured API URL and either confirms it or prompts for selection.
func (a *Service) selectAPIURL() (string, error) {
	const nuonCloudURL = "https://api.nuon.co"

	// Check if an API URL was explicitly configured (via config file or NUON_API_URL env).
	// The struct default is set directly, not via viper, so GetString returns ""
	// when no explicit value was provided.
	configuredURL := a.cfg.GetString("api_url")

	// No URL configured — show deployment type selector (first-time user)
	if configuredURL == "" {
		return a.promptDeploymentType(nuonCloudURL)
	}

	// URL is configured — show source info and confirm
	displayName := configuredURL
	if configuredURL == nuonCloudURL {
		displayName = "Nuon Cloud"
	}

	// Print dim context line showing URL and source
	source := a.cfg.APIURLSource
	fmt.Println(styles.TextDim.Render(fmt.Sprintf("  %s (%s)", configuredURL, source)))

	confirmed, err := bubbles.InlineConfirm(
		fmt.Sprintf("Login to %s", displayName),
		true,
		a.cfg.Interactive,
	)
	if err != nil {
		return "", fmt.Errorf("failed to confirm API URL: %w", err)
	}

	if confirmed {
		return configuredURL, nil
	}

	return a.promptCustomURL()
}

// promptDeploymentType shows the Nuon Cloud / Nuon BYOC selector.
func (a *Service) promptDeploymentType(nuonCloudURL string) (string, error) {
	deploymentType, err := bubbles.SelectFromOptions(
		"Which Nuon deployment are you using?",
		[]string{"Nuon Cloud", "Nuon BYOC"},
		a.cfg.Interactive,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get deployment type: %w", err)
	}

	if deploymentType == "Nuon Cloud" {
		return nuonCloudURL, nil
	}

	return a.promptCustomURL()
}

// promptCustomURL asks for a URL and normalizes the scheme.
func (a *Service) promptCustomURL() (string, error) {
	customHostname, err := bubbles.PromptText(
		"Enter your Nuon API URL:",
		"https://api.your-domain.com",
		"Example: https://api.your-domain.com or https://api.nuon.co",
		true,
		a.cfg.Interactive,
	)
	if err != nil {
		return "", fmt.Errorf("failed to get API URL: %w", err)
	}

	customHostname = strings.TrimSpace(customHostname)
	if !strings.HasPrefix(customHostname, "http://") && !strings.HasPrefix(customHostname, "https://") {
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
