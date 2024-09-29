package settings

import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

func (s *Settings) Start() error {
	if err := s.fetch(s.ctx); err != nil {
		return fmt.Errorf("unable to intialize settings: %w", err)
	}

	return nil
}

func (s *Settings) Stop() error {
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
