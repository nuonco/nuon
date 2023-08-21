package app

import (
	"fmt"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/powertoolsdev/mono/pkg/api/client"
	"github.com/powertoolsdev/mono/pkg/deprecated/api/gqlclient"
)

const (
	AUTH_TOKEN_ENV_VAR_NAME = "NUON_API_TOKEN"
	ORG_ID_VAR_NAME         = "NUON_ORG_ID"
	APP_ID_VAR_NAME         = "NUON_APP_ID"
)

type commands struct {
	v *validator.Validate

	client    client.Client
	apiClient gqlclient.Client

	appID string
	orgID string
}

// New returns a default commands with the default orgcontext getter
func New(v *validator.Validate, opts ...commandsOption) (*commands, error) {
	r := &commands{
		v: v,
	}
	for idx, opt := range opts {
		if err := opt(r); err != nil {
			return nil, fmt.Errorf("option %d failed: %w", idx, err)
		}
	}

	if err := r.v.Struct(r); err != nil {
		return nil, fmt.Errorf("unable to validate app commands: %w", err)
	}

	return r, nil
}

type commandsOption func(*commands) error

func WithAPIURL(apiURL string) commandsOption {
	return func(c *commands) error {
		authToken := os.Getenv(AUTH_TOKEN_ENV_VAR_NAME)
		if authToken == "" {
			return fmt.Errorf("No auth token set, please set $%s", AUTH_TOKEN_ENV_VAR_NAME)
		}

		gqlClient, err := gqlclient.New(c.v, gqlclient.WithAuthToken(authToken), gqlclient.WithURL(apiURL))
		if err != nil {
			return fmt.Errorf("unable to create api client")
		}
		c.apiClient = gqlClient

		// TODO: get ctl-api url from param
		ctlClient, err := client.New(c.v, client.WithAuthToken(authToken), client.WithURL(os.Getenv("NUON_API_URL")), client.WithOrgID(c.orgID))
		if err != nil {
			return fmt.Errorf("unable to create api client")
		}
		c.client = ctlClient

		return nil
	}
}

func WithDefaultEnv() commandsOption {
	return func(c *commands) error {
		c.orgID = os.Getenv(ORG_ID_VAR_NAME)
		c.appID = os.Getenv(APP_ID_VAR_NAME)
		return nil
	}
}
