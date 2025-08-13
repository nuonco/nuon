package monitor

import (
	"context"

	"go.uber.org/fx"
)

func (s *Monitor) Start() error {
	s.wg.Go(func() {
		s.loop(s.ctx)
	})
	return nil
}

func (s *Monitor) Stop() error {
	s.wg.Wait()
	return nil
}

func (s *Monitor) LifecycleHook() fx.Hook {
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
