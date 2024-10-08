package errs

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/fx"

	"github.com/powertoolsdev/mono/pkg/errs"
)

func (r *Recorder) Start() error {
	err := sentry.Init(sentry.ClientOptions{
		Dsn: errs.SentryMainDSN,
		// TODO(sdboyer): come up with a way of inferring from existing context that this is a dev build
		//Environment: r.settings.Env,
		//Tags: map[string]string{
		//"org_id": r.settings.OrgID,
		//"app":    "runner",
		//},
	})
	// It's expected that there are places the nuon binary will be executed where it is
	// not possible to connect to sentry. So we just make a note of whether sentry is active
	// for later reference.
	r.sentryEnabled = err == nil

	return nil
}

func (r *Recorder) Stop() error {
	if r.sentryEnabled {
		sentry.Flush(2 * time.Second)
	}
	return nil
}

func (r *Recorder) LifecycleHook() fx.Hook {
	return fx.Hook{
		// start the background loop to update the settings
		OnStart: func(context.Context) error {
			return r.Start()
		},

		// stop the loop and wait for the background goroutine to return
		OnStop: func(context.Context) error {
			return r.Stop()
		},
	}
}
