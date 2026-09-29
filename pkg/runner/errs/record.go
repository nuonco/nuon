package errs

import (
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/errs"
	"github.com/nuonco/nuon/pkg/runner/settings"
)

type Recorder struct {
	l             *zap.Logger
	sentryEnabled bool
	settings      *settings.Settings
}

func (r *Recorder) Record(msg string, err error) {
	r.l.Error(msg, zap.Error(err))
}

func (r *Recorder) ToSentry(err error) {
	if r.sentryEnabled {
		switch {
		case errors.Is(err, &RunnerHandlerError{}):
			errs.ReportToSentry(err, nil)
		case errors.Is(err, &RunnerFrameworkError{}):
			errs.ReportToSentry(err, nil)
		default:
			errs.ReportToSentry(WithFrameworkError(err, ""), nil)
		}
	}
}

type Params struct {
	fx.In

	L        *zap.Logger `name:"system"`
	Settings *settings.Settings
	LC       fx.Lifecycle
}

func NewRecorder(params Params) *Recorder {
	r := &Recorder{
		l:        params.L,
		settings: params.Settings,
	}

	params.LC.Append(r.LifecycleHook())
	return r
}
