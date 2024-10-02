package heartbeater

import (
	"context"

	"go.uber.org/fx"
)

func (s *HeartBeater) Start() error {
	s.wg.Go(func() {
		s.loop(s.ctx)
	})
	return nil
}

func (s *HeartBeater) Stop() error {
	s.wg.Wait()
	return nil
}

func (s *HeartBeater) LifecycleHook() fx.Hook {
	return fx.Hook{
		// start the background loop to update the settings
		OnStart: func(context.Context) error {
			s.Start()
			return nil
		},

		// stop the loop and wait for the background goroutine to return
		OnStop: func(context.Context) error {
			s.cancelFn()
			s.Stop()
			return nil
		},
	}
}
