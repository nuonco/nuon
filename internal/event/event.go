package event

import (
	"fmt"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/pkg/errors"
)

var (
	ErrWaypointJobEventNoop = fmt.Errorf("noop")
	ErrWaypointJobStream    = fmt.Errorf("error occurred querying job stream")
	ErrWaypointJobFailed    = fmt.Errorf("job failed")
)

// waypointJobStreamResponse is an interface that wraps the response and exposes methods for getting the underlying
// event.
type waypointJobStreamResponse interface {
	GetOpen() *gen.GetJobStreamResponse_Open
	GetState() *gen.GetJobStreamResponse_State
	GetJob() *gen.GetJobStreamResponse_JobChange
	GetTerminal() *gen.GetJobStreamResponse_Terminal
	GetDownload() *gen.GetJobStreamResponse_Download
	GetError() *gen.GetJobStreamResponse_Error
	GetComplete() *gen.GetJobStreamResponse_Complete
}

type waypointJobEventType = string

const (
	WaypointJobEventTypeOpen      waypointJobEventType = "open"
	WaypointJobEventTypeState     waypointJobEventType = "state"
	WaypointJobEventTypeJobChange waypointJobEventType = "job_change"
	WaypointJobEventTypeTerminal  waypointJobEventType = "terminal"
	WaypointJobEventTypeDownload  waypointJobEventType = "download"
	WaypointJobEventTypeError     waypointJobEventType = "error"
	WaypointJobEventTypeComplete  waypointJobEventType = "complete"
	WaypointJobEventTypeUnknown   waypointJobEventType = "unknown"
)

type WaypointJobEvent struct {
	Type  waypointJobEventType `json:"type"`
	JobID string               `json:"job_id"`

	// embed the actual objects in the struct so we can access the raw stream event, without having to pass the
	// stream response around
	openEv      *gen.GetJobStreamResponse_Open
	stateEv     *gen.GetJobStreamResponse_State
	jobChangeEv *gen.GetJobStreamResponse_JobChange
	terminalEv  *gen.GetJobStreamResponse_Terminal
	downloadEv  *gen.GetJobStreamResponse_Download
	errorEv     *gen.GetJobStreamResponse_Error
	completeEv  *gen.GetJobStreamResponse_Complete
}

// GetEventUnsafe returns the underlying event based on the type, with no validation
func (w WaypointJobEvent) GetEventUnsafe() interface{} {
	switch w.Type {
	case WaypointJobEventTypeOpen:
		return w.openEv
	case WaypointJobEventTypeState:
		return w.stateEv
	case WaypointJobEventTypeJobChange:
		return w.jobChangeEv
	case WaypointJobEventTypeTerminal:
		return w.terminalEv
	case WaypointJobEventTypeDownload:
		return w.downloadEv
	case WaypointJobEventTypeError:
		return w.errorEv
	case WaypointJobEventTypeComplete:
		return w.completeEv
	}

	return nil
}

// WaypointJobStreamResponseToWaypointEvent converts a stream response to an event we can handle in this package
func WaypointJobStreamResponseToWaypointEvent(jobID string, resp waypointJobStreamResponse) WaypointJobEvent {
	event := WaypointJobEvent{
		JobID: jobID,
		Type:  WaypointJobEventTypeUnknown,
	}

	if ev := resp.GetOpen(); ev != nil {
		event.Type = WaypointJobEventTypeOpen
		event.openEv = ev
	}

	if ev := resp.GetState(); ev != nil {
		event.Type = WaypointJobEventTypeState
		event.stateEv = ev
	}

	if ev := resp.GetJob(); ev != nil {
		event.Type = WaypointJobEventTypeJobChange
		event.jobChangeEv = ev
	}

	if ev := resp.GetTerminal(); ev != nil {
		event.Type = WaypointJobEventTypeTerminal
		event.terminalEv = ev
	}

	if ev := resp.GetDownload(); ev != nil {
		event.Type = WaypointJobEventTypeDownload
		event.downloadEv = ev
	}

	if ev := resp.GetError(); ev != nil {
		event.Type = WaypointJobEventTypeError
		event.errorEv = ev
	}

	if ev := resp.GetComplete(); ev != nil {
		event.Type = WaypointJobEventTypeComplete
		event.completeEv = ev
	}

	return event
}

// WaypointJobEventToErr returns whether a waypoint event was an err. If errWaypointEventNoop is returned, it
// means that no error was actually returned and that it was a standard event.
func WaypointJobEventToErr(ev WaypointJobEvent) error {
	switch ev.Type {
	case WaypointJobEventTypeError:
		return errors.Wrap(ErrWaypointJobStream, ev.errorEv.String())
	case WaypointJobEventTypeComplete:
		if ev.completeEv.Error != nil {
			return errors.Wrap(ErrWaypointJobFailed, ev.completeEv.Error.String())
		} else {
			return nil
		}
	default:
	}

	return ErrWaypointJobEventNoop
}
