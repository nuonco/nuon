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

func (j *jobLoop) Drain() {
	j.pollCancel()
}

func (j *jobLoop) Stop() error {
	j.pollCancel()
	j.jobCancel()
	j.pool.Wait()
	j.setStopped()
	return nil
}

func (j *jobLoop) LifecycleHook() fx.Hook {
	return fx.Hook{
		OnStart: func(context.Context) error {
			return j.Start()
		},
		OnStop: func(context.Context) error {
			return j.Stop()
		},
	}
}
