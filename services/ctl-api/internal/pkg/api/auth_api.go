package api

import (
	"fmt"

	"github.com/pkg/errors"
)

func NewAuthAPI(params Params) (*API, error) {
	fmt.Printf("DEBUG AUTH API - Configured Middlewares from params.Cfg.AuthMiddlewares (%d): %v\n", len(params.Cfg.AuthMiddlewares), params.Cfg.AuthMiddlewares)

	api := &API{
		cfg:                   params.Cfg,
		port:                  params.Cfg.AuthHTTPPort,
		name:                  "auth",
		services:              params.Services,
		middlewares:           params.Middlewares,
		l:                     params.L,
		configuredMiddlewares: params.Cfg.AuthMiddlewares,
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
