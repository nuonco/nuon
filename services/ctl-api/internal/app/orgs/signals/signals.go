package signals

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/eventloop"
)

const (
	TemporalNamespace string = "orgs"

	OperationCreated eventloop.SignalType = "created"
	OperationRestart eventloop.SignalType = "restart"

	OperationProvision      eventloop.SignalType = "provision"
	OperationDelete         eventloop.SignalType = "delete"
	OperationForceDelete    eventloop.SignalType = "force_delete"
	OperationDeprovision    eventloop.SignalType = "deprovision"
	OperationReprovision    eventloop.SignalType = "reprovision"
	OperationInviteCreated  eventloop.SignalType = "invite_created"
	OperationInviteAccepted eventloop.SignalType = "invite_accepted"
)

type Signal struct {
	Type eventloop.SignalType

	// fields for this event loop signal
	InviteID string

	eventloop.BaseSignal
}

var _ eventloop.Signal = (*Signal)(nil)

func (s *Signal) SignalType() eventloop.SignalType {
	return s.Type
}

func (s *Signal) Namespace() string {
	return TemporalNamespace
}

func (s *Signal) Name() string {
	return string(s.Type)
}

func (s *Signal) Start() bool {
	switch s.Type {
	case OperationCreated:
		return true
	case OperationRestart:
		return true
	default:
	}

	return false
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := middlewares.OrgFromContext(ctx)
	if err == nil {
		return org, nil
	}

	currentOrg := app.Org{}
	res := db.WithContext(ctx).
		First(&currentOrg, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get org: %w", res.Error)
	}

	return &currentOrg, nil
}
