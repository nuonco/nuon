package jobloop

import (
	"context"

	"go.uber.org/fx"
)

func (s *jobLoop) Start() error {
	s.pool.Go(s.worker)
	return nil
}

func (s *jobLoop) Stop() error {
	s.ctxCancel()
	s.pool.Wait()
	return nil
}

func (s *jobLoop) LifecycleHook() fx.Hook {
	return fx.Hook{
		// start the background loop to update the settings
		OnStart: func(context.Context) error {
			return s.Start()
		},

		// stop the loop and wait for the background goroutine to return
		OnStop: func(context.Context) error {
			return s.Stop()
		},
	}
}
