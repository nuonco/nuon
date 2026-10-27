package api

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
)

type stackRouteService interface {
	RegisterStackRoutes(*gin.Engine) error
}

func NewStackAPI(params Params) (*API, error) {
	middlewares := params.Cfg.StackMiddlewares
	if len(middlewares) == 0 {
		middlewares = params.Cfg.RunnerMiddlewares
	}
	api := &API{
		cfg:                   params.Cfg,
		port:                  params.Cfg.StackHTTPPort,
		name:                  "stack",
		services:              params.Services,
		middlewares:           params.Middlewares,
		configuredMiddlewares: middlewares,
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
		api.handler.Use(params.APISurface.Handler(permissions.APIStack))
	}
	for _, svc := range params.Services {
		stackSvc, ok := svc.(stackRouteService)
		if !ok {
			continue
		}
		if err := stackSvc.RegisterStackRoutes(api.handler); err != nil {
			return nil, errors.Wrap(err, "unable to register stack routes")
		}
	}

	params.LC.Append(api.lifecycleHooks(params.Shutdowner))
	return api, nil
}
