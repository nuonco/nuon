package settings

import (
	"context"
	"fmt"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func (s *Settings) Start() error {
	if err := s.fetch(s.ctx); err != nil {
		return fmt.Errorf("unable to intialize settings: %w", err)
	}

	s.wg.Go(func() {
		for range s.ticker.C {
			if err := s.fetch(s.ctx); err != nil {
				s.l.Error("unable to fetch settings", zap.Error(err))
			}
		}
	})
	return nil
}

func (s *Settings) Stop() error {
	s.ticker.Stop()
	s.wg.Wait()

	return nil
}

func (s *Settings) LifecycleHook() fx.Hook {
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
