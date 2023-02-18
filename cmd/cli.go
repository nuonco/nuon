package cmd

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/go-common/config"
	"github.com/powertoolsdev/nuonctl/internal"
	temporalclient "github.com/powertoolsdev/nuonctl/internal/clients/temporal"
	"github.com/powertoolsdev/nuonctl/internal/repos/executors"
	"github.com/powertoolsdev/nuonctl/internal/repos/temporal"
	"github.com/powertoolsdev/nuonctl/internal/repos/workflows"
	"github.com/spf13/pflag"
)

type cli struct {
	v *validator.Validate

	temporal  temporal.Repo
	workflows workflows.Repo
	executors executors.Repo
}

// loadConfig: load config and return it
func (c *cli) loadConfig(flags *pflag.FlagSet) (*internal.Config, error) {
	// load configuration and setup the deployments command namespace
	var cfg internal.Config
	if err := config.LoadInto(flags, &cfg); err != nil {
		return nil, fmt.Errorf("unable to load config: %w", err)
	}
	if err := cfg.Validate(c.v); err != nil {
		return nil, fmt.Errorf("unable to validate config: %w", err)
	}

	return &cfg, nil
}

func (c *cli) loadTemporalRepo(cfg *internal.Config) (temporal.Repo, error) {
	tclient, err := temporalclient.New(c.v, temporalclient.WithConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	temporal, err := temporal.New(c.v, temporal.WithClient(tclient))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	return temporal, nil
}

func (c *cli) loadWorkflowsRepo(cfg *internal.Config) (workflows.Repo, error) {
	workflows, err := workflows.New(c.v, workflows.WithConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	return workflows, nil
}

func (c *cli) loadExecutorsRepo(cfg *internal.Config) (executors.Repo, error) {
	executors, err := executors.New(c.v, executors.WithConfig(cfg))
	if err != nil {
		return nil, fmt.Errorf("unable to get temporal client: %w", err)
	}

	return executors, nil
}

func (c *cli) init(flags *pflag.FlagSet) error {
	// set config on context
	cfg, err := c.loadConfig(flags)
	if err != nil {
		return err
	}

	// set temporal on context
	temporal, err := c.loadTemporalRepo(cfg)
	if err != nil {
		return err
	}
	c.temporal = temporal

	workflows, err := c.loadWorkflowsRepo(cfg)
	if err != nil {
		return err
	}
	c.workflows = workflows

	executors, err := c.loadExecutorsRepo(cfg)
	if err != nil {
		return err
	}
	c.executors = executors

	return nil
}
