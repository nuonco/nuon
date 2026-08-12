package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/internal/agentmode"
	"github.com/nuonco/nuon/bins/cli/internal/config"
	"github.com/nuonco/nuon/bins/cli/internal/oidctoken"
	"github.com/nuonco/nuon/bins/cli/internal/services/auth"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
)

type cli struct {
	v         *validator.Validate
	apiClient nuon.Client
	ctx       context.Context
	cfg       *config.Config

	currentUserOnce sync.Once
	currentUser     *models.AppAccount
	currentUserErr  error
}

// Cached per process: the response carries every role the account holds, and
// several call sites need it.
func (c *cli) getCurrentUser(ctx context.Context) (*models.AppAccount, error) {
	c.currentUserOnce.Do(func() {
		c.currentUser, c.currentUserErr = c.apiClient.GetCurrentUser(ctx)
	})
	return c.currentUser, c.currentUserErr
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
			if err := c.tryAmbientOIDCExchange(cmd.Context()); err != nil {
				return err
			}
		}
		if c.cfg.APIToken == "" {
			return errors.New("no API token configured; run `nuon login`, set api_token in config, or configure OIDC federation (NUON_OIDC_TOKEN or GitHub Actions with `permissions: id-token: write`)")
		}
		if err := c.initUser(); err != nil {
			return errors.Wrap(err, "unable to initialize user")
		}
	}

	c.cfg.BindCobraFlags(cmd)
	return nil
}

// tryAmbientOIDCExchange exchanges an ambient OIDC token (CI) for a Nuon API
// token when no api_token is configured. The exchanged token is kept
// in-memory only: each CLI invocation exchanges fresh, so expiry never needs
// handling. Missing ambient credentials are not an error — the caller falls
// through to the standard no-token message.
func (c *cli) tryAmbientOIDCExchange(ctx context.Context) error {
	if c.cfg.OrgID == "" || !oidctoken.Available() {
		return nil
	}

	oidcJWT, source, _, err := oidctoken.Detect(ctx, oidctoken.Audience("", c.cfg.APIURL))
	if err != nil {
		return errors.Wrapf(err, "unable to get OIDC token from %s", source)
	}

	svc := auth.New(c.apiClient, c.cfg)
	resp, err := svc.ExchangeOIDCToken(ctx, oidcJWT, c.cfg.OrgID)
	if err != nil {
		return errors.Wrapf(err, "OIDC federation with token from %s failed", source)
	}

	c.cfg.APIToken = resp.Token
	if err := c.initAPIClient(); err != nil {
		return errors.Wrap(err, "unable to initialize api client with federated token")
	}

	ui.PrintLn(fmt.Sprintf("authenticated via OIDC federation (%s)", source))
	return nil
}
