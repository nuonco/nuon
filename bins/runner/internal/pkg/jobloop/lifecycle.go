package jobloop

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func (j *jobLoop) Start() error {
	loopShouldRun := generics.SliceContains(string(j.jobGroup), j.settings.Groups)
	if loopShouldRun {
		j.l.Info("should run", zap.String("group", string(j.jobGroup)))
		j.pool.Go(j.runWorker)
		j.setStarted()
	} else {
		j.l.Info("won't run", zap.String("group", string(j.jobGroup)))
	}
	return nil
}

func (j *jobLoop) Stop() error {
	loopShouldBeRunning := generics.SliceContains(string(j.jobGroup), j.settings.Groups)
	if loopShouldBeRunning {
		j.l.Info("stopping running loop", zap.String("group", string(j.jobGroup)))
		j.ctxCancel()
		j.pool.Wait()
		j.setStopped()
	} else {
		j.l.Debug("doing nothing\n", zap.String("group", string(j.jobGroup)))
	}
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
