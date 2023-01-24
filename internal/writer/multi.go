package writer

import "github.com/powertoolsdev/workers-executors/internal/event"

type multiWriter struct {
	writers []EventWriter
}

var _ EventWriter = (*multiWriter)(nil)

// NewMultiWriter creates an event writer that writes to multiple event writers
func NewMultiWriter(writers ...EventWriter) *multiWriter {
	return &multiWriter{
		writers: writers,
	}
}

// Write will write the event to the underlying event writers
func (m *multiWriter) Write(ev event.WaypointJobEvent) error {
	for _, writer := range m.writers {
		if err := writer.Write(ev); err != nil {
			return err
		}
	}
	return nil
}
