package api

import (
	"fmt"

	"github.com/pkg/errors"
)

func NewInternalAPI(params Params) (*API, error) {
	fmt.Printf("DEBUG INTERNAL API - Configured Middlewares from params.Cfg.InternalMiddlewares (%d): %v\n", len(params.Cfg.InternalMiddlewares), params.Cfg.InternalMiddlewares)

	api := &API{
		cfg:                   params.Cfg,
		port:                  params.Cfg.InternalHTTPPort,
		name:                  "internal",
		services:              params.Services,
		middlewares:           params.Middlewares,
		l:                     params.L,
		configuredMiddlewares: params.Cfg.InternalMiddlewares,
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
