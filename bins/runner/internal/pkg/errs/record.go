package errs

import (
	"errors"

	"github.com/powertoolsdev/mono/bins/runner/internal/pkg/settings"
	"github.com/powertoolsdev/mono/pkg/errs"
	"go.uber.org/fx"
	"go.uber.org/zap"
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
			errs.ReportToSentry(err)
		case errors.Is(err, &RunnerFrameworkError{}):
			errs.ReportToSentry(err)
		default:
			errs.ReportToSentry(WithFrameworkError(err, ""))
		}
	}
}

func NewRecorder(l *zap.Logger, s *settings.Settings, lc fx.Lifecycle) *Recorder {
	r := &Recorder{
		l:        l,
		settings: s,
	}

	lc.Append(r.LifecycleHook())
	return r
}
