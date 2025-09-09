package cmd

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/powertoolsdev/mono/pkg/analytics"
)

type cli struct {
	v               *validator.Validate
	apiClient       nuon.Client
	ctx             context.Context
	cfg             *config.Config
	analyticsClient analytics.Writer
}

func NewCLI() (*cli, error) {
	// Construct a validator for the API client and the UI logger.
	v := validator.New()
	c := &cli{
		v:   v,
		ctx: context.Background(),
	}

	return c, nil
}

func (c *cli) persistentPreRunE(cmd *cobra.Command, args []string) error {
	err := c.doPersistentPreRunE(cmd, args)
	if err != nil {
		// In none of the cases where this pre-run hook fails is it appropriate to print usage. But,
		// setting SilenceUsage unconditionally would cause Cobra to not print usage at some times when
		// it is appropriate.
		cmd.SilenceUsage = true
	}
	return err
}

func (c *cli) doPersistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := c.initConfig(); err != nil {
		return errors.Wrap(err, "unable to initialize config")
	}

	if err := c.initAPIClient(); err != nil {
		return errors.Wrap(err, "unable to initialize api client")
	}

	// Skip user initialization for auth commands (login, logout)
	if cmd.Use != "login" && cmd.Use != "auth" && (cmd.Parent() == nil || cmd.Parent().Use != "auth") {
		if err := c.initUser(); err != nil {
			return errors.Wrap(err, "unable to initialize user")
		}
	}

	if err := c.initSentry(); err != nil {
		return errors.Wrap(err, "unable to initialize sentry")
	}

	if err := c.initAnalytics(); err != nil {
		return errors.Wrap(err, "unable to initialize analytics")
	}

	//if err := c.checkCLIVersion(); err != nil {
	//return err
	//}

	c.cfg.BindCobraFlags(cmd)
	return nil
}
