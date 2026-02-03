package api

import "github.com/pkg/errors"

func NewAdminDashboardAPI(params Params) (*API, error) {
	api := &API{
		cfg:                   params.Cfg,
		port:                  params.Cfg.AdminDashboardHTTPPort,
		name:                  "admin-dashboard",
		services:              params.Services,
		middlewares:           params.Middlewares,
		l:                     params.L,
		configuredMiddlewares: params.Cfg.AdminDashboardMiddlewares,
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
