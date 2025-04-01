package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

type (
	EnsureEventLoopsRequest     struct{}
	EnsureEventLoopsPageRequest struct {
		Namespace string `json:"namespace"`
		Offset    int    `json:"offset"`
		Limit     int    `json:"limit"`
	}
)

const (
	TemporalNamespace string = "general"
	EventLoop         string = "general"

	OperationCreated             eventloop.SignalType = "created"
	OperationPromotion           eventloop.SignalType = "promotion"
	OperationRestart             eventloop.SignalType = "restart"
	OperationTerminateEventLoops eventloop.SignalType = "terminate_event_loops"
)

type RequestSignal struct {
	*Signal
	eventloop.EventLoopRequest
}

func NewRequestSignal(ev eventloop.EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: ev,
	}
}

type Signal struct {
	Type eventloop.SignalType `validate:"required"`

	// Only added when a promotion goes out
	Tag string `json:"tag"`

	eventloop.BaseSignal
}

var _ eventloop.Signal = (*Signal)(nil)

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}

	return nil
}

func (s *Signal) SignalType() eventloop.SignalType {
	return s.Type
}

func (s *Signal) Namespace() string {
	return TemporalNamespace
}

func (s *Signal) Name() string {
	return string(s.Type)
}

func (s *Signal) Restart() bool {
	switch s.Type {
	case OperationRestart:
		return true
	default:
	}

	return false
}

func (s *Signal) Stop() bool {
	return false
}

func (s *Signal) Start() bool {
	switch s.Type {
	case OperationCreated:
		return true
	default:
	}

	return false
}

// NOTE(fd): the general loop has no concept for an organization. we also modify startEventLoop
// to ensure the absence of an org doesn't prevent the loop from starting
func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	return nil, nil
}
