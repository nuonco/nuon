package api

import (
	"github.com/pkg/errors"
)

func NewRunnerAPI(params Params) (*API, error) {
	api := &API{
		configuredMiddlewares: params.Cfg.RunnerMiddlewares,
		cfg:                   params.Cfg,
		port:                  params.Cfg.RunnerHTTPPort,
		name:                  "runner",
		services:              params.Services,
		middlewares:           params.Middlewares,
		l:                     params.L,
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
