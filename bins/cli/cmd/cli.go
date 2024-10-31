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

func (c *cli) persistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := c.initConfig(); err != nil {
		return errors.Wrap(err, "unable to initialize config")
	}

	if err := c.initAPIClient(); err != nil {
		return errors.Wrap(err, "unable to initialize api client")
	}

	if cmd.Use != "login" {
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

	c.cfg.BindCobraFlags(cmd)
	return nil
}
