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
	"github.com/nuonco/nuon/bins/cli/internal/services/actions"
	"github.com/nuonco/nuon/bins/cli/internal/services/apps"
	"github.com/nuonco/nuon/bins/cli/internal/services/auth"
	"github.com/nuonco/nuon/bins/cli/internal/services/builds"
	"github.com/nuonco/nuon/bins/cli/internal/services/components"
	"github.com/nuonco/nuon/bins/cli/internal/services/docs"
	"github.com/nuonco/nuon/bins/cli/internal/services/installs"
	"github.com/nuonco/nuon/bins/cli/internal/services/mcpserver"
	"github.com/nuonco/nuon/bins/cli/internal/services/orgs"
	"github.com/nuonco/nuon/bins/cli/internal/services/roles"
	"github.com/nuonco/nuon/bins/cli/internal/services/runbooks"
	"github.com/nuonco/nuon/bins/cli/internal/services/secrets"
	"github.com/nuonco/nuon/bins/cli/internal/services/serviceaccounts"
	"github.com/nuonco/nuon/bins/cli/internal/services/triggers"
	"github.com/nuonco/nuon/bins/cli/internal/services/variables"
	"github.com/nuonco/nuon/bins/cli/internal/services/version"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/sdks/stack/oidctoken"
)

type cli struct {
	v         *validator.Validate
	apiClient nuon.Client
	ctx       context.Context
	cfg       *config.Config

	actions         *actions.Service
	apps            *apps.Service
	auth            *auth.Service
	builds          *builds.Service
	components      *components.Service
	docs            *docs.Service
	installs        *installs.Service
	mcpserver       *mcpserver.Service
	orgs            *orgs.Service
	roles           *roles.Service
	runbooks        *runbooks.Service
	secrets         *secrets.Service
	serviceAccounts *serviceaccounts.Service
	triggers        *triggers.Service
	variables       *variables.Service
	version         *version.Service

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

	// The auth token must be resolved before the fx graph is built: services
	// capture the API client at construction, so a client built with an empty
	// token cannot be swapped out afterwards.
	skipAuth := hasSkipAuthAnnotation(cmd)
	if !skipAuth {
		if c.cfg.APIToken == "" {
			if err := c.tryAmbientOIDCExchange(cmd.Context()); err != nil {
				return err
			}
		}
		if c.cfg.APIToken == "" {
			return errors.New("no API token configured; run `nuon login`, set api_token in config, or configure OIDC federation (NUON_OIDC_TOKEN or GitHub Actions with `permissions: id-token: write`)")
		}
	}

	if err := c.populateDeps(); err != nil {
		return errors.Wrap(err, "unable to initialize CLI dependencies")
	}

	if !skipAuth {
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

	// The exchange runs before the fx graph exists, so it needs its own
	// (tokenless) client: the exchange endpoint is unauthenticated.
	exchangeClient, err := newAPIClient(c.v, c.cfg)
	if err != nil {
		return errors.Wrap(err, "unable to initialize api client for OIDC exchange")
	}

	svc := auth.New(exchangeClient, c.cfg)
	resp, err := svc.ExchangeOIDCToken(ctx, oidcJWT, c.cfg.OrgID)
	if err != nil {
		return errors.Wrapf(err, "OIDC federation with token from %s failed", source)
	}

	c.cfg.APIToken = resp.Token
	ui.PrintLn(fmt.Sprintf("authenticated via OIDC federation (%s)", source))
	return nil
}
