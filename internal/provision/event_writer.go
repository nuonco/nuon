package provision

import (
	"github.com/powertoolsdev/go-waypoint/job"
)

type logEventWriter struct {
	log logger
}

type logger interface {
	Info(string, ...interface{})
}

func newLogEventWriter(l logger) *logEventWriter {
	return &logEventWriter{
		log: l,
	}
}

func (l logEventWriter) Write(event job.WaypointJobEvent) error {
	l.log.Info("job-event: jobID=%s type=%s event=%v", event.JobID, event.Type, event)
	return nil
}
