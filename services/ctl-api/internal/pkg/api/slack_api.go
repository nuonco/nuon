package api

import (
	"github.com/pkg/errors"
)

// NewSlackAPI builds the dedicated HTTP listener for Slack-originating
// traffic: OAuth callback, slash commands, and Slack Events API webhooks.
//
// This is a separate listener (its own port, its own ingress) rather than
// being mounted on the public API to keep Slack's signing-secret-based auth
// model isolated from the API-key + org-id model used by /v1/*.
func NewSlackAPI(params Params) (*API, error) {
	api := &API{
		cfg:                   params.Cfg,
		port:                  params.Cfg.SlackHTTPPort,
		name:                  "slack",
		services:              params.Services,
		middlewares:           params.Middlewares,
		configuredMiddlewares: params.Cfg.SlackMiddlewares,
		l:                     params.L,
		db:                    params.DB,
		endpointAudit:         params.EndpointAudit,
	}

	if err := api.init(); err != nil {
		return nil, errors.Wrap(err, "unable to initialize")
	}

	if err := api.registerMiddlewares(); err != nil {
		return nil, errors.Wrap(err, "unable to register middlewares")
	}

	if err := api.registerServices(); err != nil {
		return nil, errors.Wrap(err, "unable to register services")
	}

	params.LC.Append(api.lifecycleHooks(params.Shutdowner))
	return api, nil
}
