package job

// EventWriter allows for writing of a waypoint event into a shared buffer / stream
type EventWriter interface {
	Write(WaypointJobEvent) error
}

// multiWriter allows for multiple EventWriters so we can write waypoint events in multiple streams
type multiWriter struct {
	writers []EventWriter
}

func NewMultiWriter(writers ...EventWriter) multiWriter {
	return multiWriter{
		writers: writers,
	}
}

func (m multiWriter) Write(ev WaypointJobEvent) error {
	for _, writer := range m.writers {
		if err := writer.Write(ev); err != nil {
			return err
		}
	}
	return nil
}
