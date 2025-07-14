package jobloop

import (
	"context"

	"github.com/powertoolsdev/mono/pkg/generics"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// NOTE: j.jobGroup may be singular or plural and is not consistent w/ the consts for RunnerJobGroups
// as a result, we use this pre-processor to make j.jobGroup and the strings in Groups plural so the
// membership check works in the db, the Groups for the runner settings look like this:
// - org: {operations,sync,build,sandbox,runner}
// - install: {operations,sync,deploys,action,sandbox}
func jobLoopShouldRun(jobGroup string, groups []string) bool {
	// 1. map the jog group to its RunnerJobGroupType const
	normalizedJobGroup := jobGroup
	switch jobGroup {
	case "builds":
		normalizedJobGroup = "build"
	case "actions":
		normalizedJobGroup = "action"
	default:
		normalizedJobGroup = jobGroup
	}
	return generics.SliceContains(normalizedJobGroup, groups)
}

func (j *jobLoop) Start() error {
	loopShouldRun := jobLoopShouldRun(string(j.jobGroup), j.settings.Groups)
	if loopShouldRun {
		j.l.Info("should run this job loop", zap.String("group", string(j.jobGroup)))
		j.pool.Go(j.runWorker)
		j.setStarted()
	} else {
		j.l.Info("should not run this job loop", zap.String("group", string(j.jobGroup)))
	}
	return nil
}

func (j *jobLoop) Stop() error {
	loopShouldBeRunning := jobLoopShouldRun(string(j.jobGroup), j.settings.Groups)
	if loopShouldBeRunning {
		j.l.Info("stopping running job loop", zap.String("group", string(j.jobGroup)))
		j.ctxCancel()
		j.pool.Wait()
		j.setStopped()
	} else {
		j.l.Debug("no running job loop", zap.String("group", string(j.jobGroup)))
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
