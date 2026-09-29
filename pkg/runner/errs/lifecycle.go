package errs

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
	"go.uber.org/fx"

	"github.com/nuonco/nuon/pkg/errs"
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
		OnStart: func(context.Context) error {
			return r.Start()
		},

		OnStop: func(context.Context) error {
			return r.Stop()
		},
	}
}
