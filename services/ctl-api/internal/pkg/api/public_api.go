package api

import (
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/org"
	"github.com/pkg/errors"
)

func NewPublicAPI(params Params) (*API, error) {
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

	if err := org.ValidateRouteCoverage(api.handler.Routes()); err != nil {
		return nil, errors.Wrap(err, "resource-grant route coverage validation failed")
	}

	params.LC.Append(api.lifecycleHooks(params.Shutdowner))
	return api, nil
}
