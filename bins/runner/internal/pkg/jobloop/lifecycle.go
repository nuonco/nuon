package jobloop

import (
	"context"
	"time"

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

func (j *jobLoop) Drain(timeout time.Duration) {
	close(j.drainCh)

	done := make(chan struct{})
	go func() {
		j.pool.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		j.ctxCancel()
		j.pool.Wait()
	}
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
