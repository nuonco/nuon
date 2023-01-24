package poll

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/waypoint/pkg/server/gen"
	"github.com/powertoolsdev/workers-executors/internal/event"
	"google.golang.org/grpc"
)

type jobStreamGetter interface {
	GetJobStream(context.Context, *gen.GetJobStreamRequest, ...grpc.CallOption) (gen.Waypoint_GetJobStreamClient, error)
}

var _ jobStreamGetter = (gen.WaypointClient)(nil)

// eventWriter allows for writing of a waypoint event into a shared buffer / stream
type eventWriter interface {
	Write(event.WaypointJobEvent) error
}

type poller struct {
	Client jobStreamGetter `validate:"required"`
	Writer eventWriter     `validate:"required"`
	JobID  string          `validate:"required"`

	// internal state
	v *validator.Validate
}

type pollerOption func(*poller) error

func New(v *validator.Validate, opts ...pollerOption) (*poller, error) {
	p := &poller{v: v}

	if v == nil {
		return nil, fmt.Errorf("error instantiating poll task: validator is nil")
	}

	for _, opt := range opts {
		if err := opt(p); err != nil {
			return nil, err
		}
	}

	if err := p.v.Struct(p); err != nil {
		return nil, err
	}

	return p, nil
}

func WithClient(c jobStreamGetter) pollerOption {
	return func(p *poller) error {
		p.Client = c
		return nil
	}
}

func WithWriter(w eventWriter) pollerOption {
	return func(p *poller) error {
		p.Writer = w
		return nil
	}
}

func WithJobID(id string) pollerOption {
	return func(p *poller) error {
		p.JobID = id
		return nil
	}
}

// Poll polls a job to completion or error , writing events to the provided event writer func
func (p *poller) Poll(ctx context.Context) error {
	streamClient, err := p.Client.GetJobStream(ctx, &gen.GetJobStreamRequest{
		JobId: p.JobID,
	})
	if err != nil {
		return fmt.Errorf("unable to get job stream: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		default:
		}

		var resp *gen.GetJobStreamResponse
		resp, err := streamClient.Recv()
		if err != nil {
			return fmt.Errorf("error while receiving response: %w", err)
		}

		// handle err
		ev := event.WaypointJobStreamResponseToWaypointEvent(p.JobID, resp)
		if err := p.Writer.Write(ev); err != nil {
			return err
		}

		wpErr := event.WaypointJobEventToErr(ev)
		switch wpErr {
		case event.ErrWaypointJobEventNoop:
			continue
		case nil:
			return nil
		default:
			return wpErr
		}
	}
}
