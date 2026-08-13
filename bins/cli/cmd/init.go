package cmd

import (
	"fmt"
	stdhttp "net/http"

	"github.com/cockroachdb/errors"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/httpdebug"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
)

// Construct an API client for the services to use.
func newAPIClient(v *validator.Validate, cfg *config.Config) (nuon.Client, error) {
	var transport stdhttp.RoundTripper
	if Debug {
		transport = httpdebug.NewTransport(nil)
	}

	api, err := nuon.New(
		nuon.WithValidator(v),
		nuon.WithAuthToken(cfg.APIToken),
		nuon.WithOrgID(cfg.OrgID),
		nuon.WithURL(cfg.APIURL),
		nuon.WithHTTPTransport(transport),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to init API client: %w", err)
	}
	api.SetClientVersion(version.Version)

	return api, nil
}

func (c *cli) initAPIClient() error {
	api, err := newAPIClient(c.v, c.cfg)
	if err != nil {
		return err
	}

	c.apiClient = api
	return nil
}

func (c *cli) initConfig() error {
	cfg, err := config.NewConfig(ConfigFile)
	if err != nil {
		return fmt.Errorf("unable to initialize config: %w", err)
	}

	c.cfg = cfg
	return nil
}

func (c *cli) initUser() error {
	if c.cfg.APIToken == "" {
		return nil
	}
	user, err := c.getCurrentUser(c.ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get current user")
	}

	c.cfg.UserID = user.ID
	return nil
}
