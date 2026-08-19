package registry

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func (s *Registry) LifecycleHook() fx.Hook {
	return fx.Hook{
		// start the background loop to update the settings
		OnStart: func(ctx context.Context) error {
			s.wg.Go(func() {
				if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					s.l.Error("embedded registry stopped unexpectedly", zap.Error(err))
					if err := s.shutdown.Shutdown(fx.ExitCode(1)); err != nil {
						s.l.Warn("unable to request runner shutdown", zap.Error(err))
					}
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
