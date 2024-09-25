package jobloop

import (
	"context"
	"fmt"

	"github.com/sourcegraph/conc/panics"
	"github.com/sourcegraph/conc/pool"
	"go.uber.org/fx"
)

func (s *jobLoop) panicWrapper(fn func()) func() {
	return func() {
		var pc panics.Catcher
		pc.Try(fn)
		recovered := pc.Recovered()

		fmt.Println("job execution: \n" + string(recovered.Stack))
	}
}

func (s *jobLoop) Start() error {
	s.pool = pool.New().WithMaxGoroutines(1)
	s.pool.Go(s.panicWrapper(s.worker))
	return nil
}

func (s *jobLoop) Stop() error {
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
