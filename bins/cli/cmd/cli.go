package cmd

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon-go"
	"github.com/powertoolsdev/mono/bins/cli/internal/config"
	"github.com/spf13/cobra"
)

type cli struct {
	v         *validator.Validate
	apiClient nuon.Client
	ctx       context.Context
	cfg       *config.Config
}

func (c *cli) persistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := c.initConfig(); err != nil {
		return err
	}

	if err := c.initAPIClient(); err != nil {
		return err
	}

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
