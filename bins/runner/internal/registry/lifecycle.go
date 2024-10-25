package registry

import (
	"context"
	"fmt"
	"log"

	"go.uber.org/fx"
)

func (s *Registry) LifecycleHook() fx.Hook {
	return fx.Hook{
		// start the background loop to update the settings
		OnStart: func(ctx context.Context) error {
			s.wg.Go(func() {
				if err := s.ListenAndServe(); err != nil {
					log.Fatal(err)
				}
			})

			return nil
		},

		// stop the loop and wait for the background goroutine to return
		OnStop: func(ctx context.Context) error {
			if err := s.Shutdown(ctx); err != nil {
				return fmt.Errorf("unable to shut down registry: %w", err)
			}

			s.cancelFn()
			s.wg.Wait()
			return nil
		},
	}
}
