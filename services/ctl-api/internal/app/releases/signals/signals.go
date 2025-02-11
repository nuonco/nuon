package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/cctx"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "releases"
)

const (
	OperationProvision        eventloop.SignalType = "provision"
	OperationCreated          eventloop.SignalType = "created"
	OperationRestart          eventloop.SignalType = "restart"
	OperationPollDependencies eventloop.SignalType = "poll_dependencies"
)

type Signal struct {
	Type eventloop.SignalType `validate:"required"`

	eventloop.BaseSignal
}

type RequestSignal struct {
	*Signal
	eventloop.EventLoopRequest
}

func NewRequestSignal(req eventloop.EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

func (s *Signal) SignalType() eventloop.SignalType {
	return s.Type
}

func (s *Signal) Stop() bool {
	return false
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

func (s *Signal) Start() bool {
	switch s.Type {
	case OperationCreated:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err == nil {
		return org, nil
	}

	cmpRelease := app.ComponentRelease{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&cmpRelease, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component release: %w", res.Error)
	}

	return &cmpRelease.Org, nil
}
