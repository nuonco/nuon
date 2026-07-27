package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/analytics"
)

type cli struct {
	v               *validator.Validate
	apiClient       nuon.Client
	ctx             context.Context
	cfg             *config.Config
	analyticsClient analytics.Writer

	org     *models.AppOrg
	orgInit bool
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

// resolveOutput determines the output format from flags and env, then enables
// json/agent modes accordingly. Precedence: --output, --json, NUON_OUTPUT,
// NUON_AGENT, default (table).
func (c *cli) resolveOutput(cmd *cobra.Command) error {
	out := "table"
	switch {
	case cmd.Flags().Changed("output"):
		out = Output
	case cmd.Flags().Changed("json"):
		out = "json"
	case os.Getenv("NUON_OUTPUT") != "":
		out = os.Getenv("NUON_OUTPUT")
	case agentmode.FromEnv():
		out = "agent"
	}

	out = strings.ToLower(strings.TrimSpace(out))
	switch out {
	case OutputTable:
	case OutputJSON:
		PrintJSON = true
		ui.SetJSONOutput(true)
	case OutputAgent:
		PrintJSON = true
		agentmode.SetEnabled(true)
	default:
		return &ui.CLIUserError{Msg: fmt.Sprintf("invalid --output %q: must be one of table, json, agent", out)}
	}

	if !supportsOutput(cmd, out) {
		return &ui.CLIUserError{Msg: fmt.Sprintf("`%s` does not support --output %s (supported: %s)", cmd.CommandPath(), out, strings.Join(supportedOutputs(cmd), ", "))}
	}

	Output = out
	return nil
}

func (c *cli) doPersistentPreRunE(cmd *cobra.Command, args []string) error {
	if err := c.resolveOutput(cmd); err != nil {
		return err
	}

	if err := guardReadOnly(cmd); err != nil {
		return err
	}

	if err := c.initConfig(); err != nil {
		return errors.Wrap(err, "unable to initialize config")
	}
	if agentmode.Enabled() {
		c.cfg.Interactive = false
	}

	if err := c.initAPIClient(); err != nil {
		return errors.Wrap(err, "unable to initialize api client")
	}

	// Skip user initialization for auth commands (login, logout)
	if !hasSkipAuthAnnotation(cmd) {
		if c.cfg.APIToken == "" {
			return errors.New("no API token configured; run `nuon login` or set api_token in config")
		}
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
