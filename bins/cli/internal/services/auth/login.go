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

type LoginResult struct {
	AccessToken string
	DisplayName string
}

func (a *Service) Login(ctx context.Context) error {
	apiURL, err := a.selectAPIURL()
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't select API URL: %w", err))
	}

	a.cfg.Set("api_url", apiURL)
	a.cfg.APIURL = apiURL

	if err := a.updateAPIClient(apiURL, a.cfg); err != nil {
		return ui.PrintError(fmt.Errorf("couldn't update API client: %w", err))
	}

	cfg, err := a.api.GetCLIConfig(ctx)
	if err != nil {
		return ui.PrintError(fmt.Errorf("couldn't get cli config: %w", err))
	}

	result, err := a.loginWithNuonAuth(ctx, cfg)
	if err != nil {
		return ui.PrintError(err)
	}

	a.cfg.Set("api_token", result.AccessToken)
	if err := a.cfg.WriteConfig(); err != nil {
		return ui.PrintError(err)
	}

	ui.PrintLn(fmt.Sprintf("Logged in as %s", result.DisplayName))

	a.printVersionNotice(ctx)

	api, err := nuon.New(
		nuon.WithValidator(validator.New()),
		nuon.WithAuthToken(result.AccessToken),
		nuon.WithURL(a.cfg.APIURL),
	)
	if err != nil {
		return ui.PrintError(fmt.Errorf("unable to init API client: %w", err))
	}
	a.api = api

	orgs, _, err := a.api.GetOrgs(ctx, &models.GetPaginatedQuery{
		Offset: 0,
		Limit:  10,
	})
	if err != nil {
		return ui.PrintError(err)
	}

	switch len(orgs) {
	case 0:
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

func (a *Service) selectAPIURL() (string, error) {
	const nuonCloudURL = "https://api.nuon.co"

	configuredURL := a.cfg.GetString("api_url")

	if configuredURL == "" {
		return a.promptDeploymentType(nuonCloudURL)
	}

	displayName := configuredURL
	if configuredURL == nuonCloudURL {
		displayName = "Nuon Cloud"
	}

	source := a.cfg.APIURLSource
	fmt.Println(styles.TextDim.Render(fmt.Sprintf("  %s (%s)", configuredURL, source)))

	confirmed, err := bubbles.InlineConfirm(
		fmt.Sprintf("Login to %s", displayName),
		true,
		a.cfg.Interactive,
	)
	if err != nil {
		if !a.cfg.Interactive {
			return "", fmt.Errorf("logging in requires an interactive terminal")
		}
		return "", fmt.Errorf("failed to confirm API URL: %w", err)
	}

	if confirmed {
		return configuredURL, nil
	}

	return a.promptCustomURL()
}

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

func (a *Service) updateAPIClient(apiURL string, cliCfg *config.Config) error {
	v := validator.New()

	api, err := nuon.New(
		nuon.WithValidator(v),
		nuon.WithAuthToken(""),
		nuon.WithOrgID(""),
		nuon.WithURL(apiURL),
	)
	if err != nil {
		return fmt.Errorf("unable to create API client with URL %s: %w", apiURL, err)
	}

	a.api = api

	return nil
}
