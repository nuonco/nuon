package api

import (
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
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
	if err := api.init(params.HTTPMetrics); err != nil {
		return nil, errors.Wrap(err, "unable to initialize")
	}

	if err := api.registerMiddlewares(); err != nil {
		return nil, errors.Wrap(err, "unable to register middlewares")
	}
	if params.InstallAuth != nil {
		api.handler.Use(params.InstallAuth.Handler())
	}
	if params.APISurface != nil {
		api.handler.Use(params.APISurface.Handler(permissions.APICtl))
	}

	if err := api.registerServices(); err != nil {
		return nil, errors.Wrap(err, "unable to register middlewares")
	}

	params.LC.Append(api.lifecycleHooks(params.Shutdowner))
	return api, nil
}
