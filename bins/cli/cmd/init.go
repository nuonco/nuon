package cmd

import (
	"fmt"
	stdhttp "net/http"

	"github.com/cockroachdb/errors"
	"github.com/getsentry/sentry-go"
	"github.com/nuonco/nuon/sdks/nuon-go"

	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/httpdebug"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/pkg/errs"
)

// Construct an API client for the services to use.
func (c *cli) initAPIClient() error {
	var transport stdhttp.RoundTripper
	if Debug {
		transport = httpdebug.NewTransport(nil)
	}

	api, err := nuon.New(
		nuon.WithValidator(c.v),
		nuon.WithAuthToken(c.cfg.APIToken),
		nuon.WithOrgID(c.cfg.OrgID),
		nuon.WithURL(c.cfg.APIURL),
		nuon.WithHTTPTransport(transport),
	)
	if err != nil {
		return fmt.Errorf("unable to init API client: %w", err)
	}
	api.SetClientVersion(version.Version)

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

func (c *cli) initSentry() error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         c.cfg.SentryDSN,
		Environment: c.cfg.Env,
		Tags: map[string]string{
			"org_id":   c.cfg.OrgID,
			"platform": "cli",
		},
	})

	if err != nil {
		wrappedErr := errors.Wrap(err, "unable to initialize sentry")
		errs.ReportToSentry(wrappedErr, nil)
		return wrappedErr
	}

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
