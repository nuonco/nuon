package queue

import (
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/activities"
)

const (
	StatusHandlerName string = "status"
	StatusHandlerType        = handlerTypeUpdate
)

type StatusRequest struct{}

type StatusResponse struct {
	Ready   bool
	Stopped bool
	Paused  bool

	QueueDepthCount int
	InFlightCount   int
	InFlight        []string
}

func (w *queue) statusHandler(ctx workflow.Context, req *StatusRequest) (*StatusResponse, error) {
	resp := &StatusResponse{
		Ready:   w.ready,
		Stopped: w.stopped,
		Paused:  w.paused,
	}
	if !w.ready {
		return resp, nil
	}

	resp.QueueDepthCount = w.ch.Len()

	queueSignals, err := activities.AwaitGetQueueSignalsByQueueID(ctx, w.queueID)
	if err != nil {
		return nil, err
	}

	resp.InFlightCount = len(queueSignals)
	resp.InFlight = make([]string, 0, len(queueSignals))
	for _, qs := range queueSignals {
		resp.InFlight = append(resp.InFlight, qs.ID)
	}

	return resp, nil
}
