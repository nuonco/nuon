package job

import (
	"fmt"

	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/pkg/errors"
)

var (
	errWaypointJobEventNoop = fmt.Errorf("noop")
	errWaypointJobStream    = fmt.Errorf("error occurred querying job stream")
	errWaypointJobFailed    = fmt.Errorf("job failed")
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
	waypointJobEventTypeOpen      waypointJobEventType = "open"
	waypointJobEventTypeState     waypointJobEventType = "state"
	waypointJobEventTypeJobChange waypointJobEventType = "job_change"
	waypointJobEventTypeTerminal  waypointJobEventType = "terminal"
	waypointJobEventTypeDownload  waypointJobEventType = "download"
	waypointJobEventTypeError     waypointJobEventType = "error"
	waypointJobEventTypeComplete  waypointJobEventType = "complete"
	waypointJobEventTypeUnknown   waypointJobEventType = "unknown"
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
	case waypointJobEventTypeOpen:
		return w.openEv
	case waypointJobEventTypeState:
		return w.stateEv
	case waypointJobEventTypeJobChange:
		return w.jobChangeEv
	case waypointJobEventTypeTerminal:
		return w.terminalEv
	case waypointJobEventTypeDownload:
		return w.downloadEv
	case waypointJobEventTypeError:
		return w.errorEv
	case waypointJobEventTypeComplete:
		return w.completeEv
	}

	return nil
}

// convert a stream response to an event we can handle in this package
func waypointJobStreamResponseToWaypointEvent(jobID string, resp waypointJobStreamResponse) WaypointJobEvent {
	event := WaypointJobEvent{
		JobID: jobID,
		Type:  waypointJobEventTypeUnknown,
	}

	if ev := resp.GetOpen(); ev != nil {
		event.Type = waypointJobEventTypeOpen
		event.openEv = ev
	}

	if ev := resp.GetState(); ev != nil {
		event.Type = waypointJobEventTypeState
		event.stateEv = ev
	}

	if ev := resp.GetJob(); ev != nil {
		event.Type = waypointJobEventTypeJobChange
		event.jobChangeEv = ev
	}

	if ev := resp.GetTerminal(); ev != nil {
		event.Type = waypointJobEventTypeTerminal
		event.terminalEv = ev
	}

	if ev := resp.GetDownload(); ev != nil {
		event.Type = waypointJobEventTypeDownload
		event.downloadEv = ev
	}

	if ev := resp.GetError(); ev != nil {
		event.Type = waypointJobEventTypeError
		event.errorEv = ev
	}

	if ev := resp.GetComplete(); ev != nil {
		event.Type = waypointJobEventTypeComplete
		event.completeEv = ev
	}

	return event
}

// waypointJobEventToErr returns whether a waypoint event was an err. If errWaypointEventNoop is returned, it
// means that no error was actually returned and that it was a standard event.
func waypointJobEventToErr(ev WaypointJobEvent) error {
	switch ev.Type {
	case waypointJobEventTypeError:
		return errors.Wrap(errWaypointJobStream, ev.errorEv.String())
	case waypointJobEventTypeComplete:
		if ev.completeEv.Error != nil {
			return errors.Wrap(errWaypointJobFailed, ev.completeEv.Error.String())
		} else {
			return nil
		}
	default:
	}

	return errWaypointJobEventNoop
}
