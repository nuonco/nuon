package errs

import "go.uber.org/zap"

type Recorder struct {
	l *zap.Logger
}

// Record is used to record errors that can not other wise be handled, such as when doing cleanup work, where we only
// care about what failed, but not the cleanup step.
func (r *Recorder) Record(msg string, err error) {
	r.l.Error(msg, zap.Error(err))
}

func NewRecorder(l *zap.Logger) *Recorder {
	return &Recorder{
		l: l,
	}
}
