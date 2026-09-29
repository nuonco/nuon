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
		OnStart: func(context.Context) error {
			s.Start()
			return nil
		},

		OnStop: func(context.Context) error {
			s.cancelFn()
			s.Stop()
			return nil
		},
	}
}
