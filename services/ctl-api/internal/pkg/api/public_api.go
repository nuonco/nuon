package api

import (
	"fmt"

	"github.com/pkg/errors"
)

func NewPublicAPI(params Params) (*API, error) {
	fmt.Printf("DEBUG PUBLIC API - Configured Middlewares from params.Cfg.Middlewares (%d): %v\n", len(params.Cfg.Middlewares), params.Cfg.Middlewares)

	api := &API{
		configuredMiddlewares: params.Cfg.Middlewares,
		cfg:                   params.Cfg,
		port:                  params.Cfg.HTTPPort,
		name:                  "public",
		services:              params.Services,
		middlewares:           params.Middlewares,
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
		return nil, errors.Wrap(err, "unable to register middlewares")
	}

	params.LC.Append(api.lifecycleHooks(params.Shutdowner))
	return api, nil
}
