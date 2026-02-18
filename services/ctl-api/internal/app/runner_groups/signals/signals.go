package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "runner-groups"
	EventLoop         string = "runner-groups"

	OperationCreated     eventloop.SignalType = "created"
	OperationRestart     eventloop.SignalType = "restart"
	OperationElectLeader eventloop.SignalType = "elect_leader"
	OperationSetLeader   eventloop.SignalType = "set_leader"
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

	// RequestedLeaderRunnerID is used with OperationSetLeader to explicitly select a leader runner.
	RequestedLeaderRunnerID string `json:"requested_leader_runner_id,omitempty"`

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
	case OperationCreated, OperationElectLeader, OperationSetLeader:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	var group app.RunnerGroup
	res := db.WithContext(ctx).Select("org_id").First(&group, "id = ? AND deleted_at = 0", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner group: %w", res.Error)
	}

	var org app.Org
	res = db.WithContext(ctx).First(&org, "id = ?", group.OrgID)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &org, nil
}
