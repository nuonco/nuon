package jobloop

import (
	"context"

	"go.uber.org/fx"
)

func (j *jobLoop) Start() error {
	j.setStarted()
	j.pool.Go(j.runWorker)
	return nil
}

func (j *jobLoop) Stop() error {
	j.ctxCancel()
	j.pool.Wait()
	j.setStopped()
	return nil
}

func (j *jobLoop) LifecycleHook() fx.Hook {
	return fx.Hook{
		// start the background loop to update the settings
		OnStart: func(context.Context) error {
			return j.Start()
		},

		// stop the loop and wait for the background goroutine to return
		OnStop: func(context.Context) error {
			return j.Stop()
		},
	}
}
