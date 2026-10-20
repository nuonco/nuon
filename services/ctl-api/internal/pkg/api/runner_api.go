package api

import (
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

func NewRunnerAPI(params Params) (*API, error) {
	api := &API{
		cfg:                   params.Cfg,
		port:                  params.Cfg.RunnerHTTPPort,
		name:                  "runner",
		services:              params.Services,
		middlewares:           params.Middlewares,
		configuredMiddlewares: params.Cfg.RunnerMiddlewares,
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
	if params.APISurface != nil {
		api.handler.Use(params.APISurface.Handler(permissions.APIRunner))
	}
	if params.RunnerGuard != nil {
		api.handler.Use(params.RunnerGuard.JobHandler())
		api.handler.Use(params.RunnerGuard.InstallHandler())
	}

	if err := api.registerServices(); err != nil {
		return nil, errors.Wrap(err, "unable to register middlewares")
	}

	params.LC.Append(api.lifecycleHooks(params.Shutdowner))
	return api, nil
}
