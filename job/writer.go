package job

// EventWriter allows for writing of a waypoint event into a shared buffer / stream
type EventWriter interface {
	Write(WaypointJobEvent) error
}
