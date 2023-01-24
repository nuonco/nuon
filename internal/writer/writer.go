package writer

import "github.com/powertoolsdev/workers-executors/internal/event"

// EventWriter allows for writing of a waypoint event into a shared buffer / stream
type EventWriter interface {
	Write(event.WaypointJobEvent) error
}
