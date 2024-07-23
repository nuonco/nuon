package cmd

import (
	"context"
	"fmt"

	"github.com/getsentry/sentry-go"
	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/pkg/errs"
	"github.com/spf13/cobra"
)

type cli struct {
	v         *validator.Validate
	apiClient nuon.Client
	ctx       context.Context
	cfg       *config.Config
	err       error
	useSentry bool
}

func (c *cli) persistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := c.initConfig(); err != nil {
		return err
	}

	if err := c.initAPIClient(); err != nil {
		return err
	}

	c.initSentry()

	c.cfg.BindCobraFlags(cmd)
	return nil
}

// Construct an API client for the services to use.
func (c *cli) initAPIClient() error {
	api, err := nuon.New(
		nuon.WithValidator(c.v),
		nuon.WithAuthToken(c.cfg.APIToken),
		nuon.WithOrgID(c.cfg.OrgID),
		nuon.WithURL(c.cfg.APIURL),
	)
	if err != nil {
		return fmt.Errorf("unable to init API client: %w", err)
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

func (c *cli) initSentry() {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: errs.SentryMainDSN,
		// TODO(sdboyer): come up with a way of inferring from existing context that this is a dev build
		Environment: "prod",
		Tags: map[string]string{
			"org_id": c.cfg.OrgID,
			"app":    "cli",
		},
	})
	// It's expected that there are places the nuon binary will be executed where it is
	// not possible to connect to sentry. So we just make a note of whether sentry is active
	// for later reference.
	c.useSentry = err == nil
}

type cobraRunCommand func(*cobra.Command, []string)
type cobraRunECommand func(*cobra.Command, []string) error

// run wraps all CLI commands, providing a central point to control error flow and handling.
func (c *cli) run(f cobraRunECommand) cobraRunCommand {
	return func(cmd *cobra.Command, args []string) {
		c.err = f(cmd, args)
		if c.err != nil {
			errs.ReportToSentry(c.err)
		}
	}
}
