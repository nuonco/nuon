package errs

import (
	"errors"

	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/errs"
)

type Recorder struct {
	l             *zap.Logger
	sentryEnabled bool
	settings      *settings.Settings
}

// Record is used to record errors that can not other wise be handled, such as when doing cleanup work, where we only
// care about what failed, but not the cleanup step.
func (r *Recorder) Record(msg string, err error) {
	r.l.Error(msg, zap.Error(err))
}

func (r *Recorder) ToSentry(err error) {
	if r.sentryEnabled {
		switch {
		// this is probably the right way - unwrap errors all the way down to see if they're one of our types,
		// and only rewrap it if it's not one of those types
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
