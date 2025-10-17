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
	TemporalNamespace string = "app-branches"

	OperationCreated eventloop.SignalType = "created"
	OperationRestart eventloop.SignalType = "restart"

	OperationCheckChanges    eventloop.SignalType = "check-changes"
	OperationUpdateAppConfig eventloop.SignalType = "update-app-config"
	OperationBuildComponents eventloop.SignalType = "build-components"
	OperationUpdateInstalls  eventloop.SignalType = "update-installs"

	// the following will be sent to a different namespace
	OperationExecuteFlow eventloop.SignalType = "execute-flow"
	OperationRerunFlow   eventloop.SignalType = "rerun-flow"
)

type Signal struct {
	Type eventloop.SignalType `json:"type"`

	AppBranchID string `json:"app_branch_id"`

	// WorkflowID is method in base signals
	FlowID           string `json:"workflow_id"`
	WorkflowStepID   string `json:"workflow_step_id"`
	WorkflowStepName string `json:"workflow_step_name"`

	eventloop.BaseSignal
}

func NewRequestSignal(req eventloop.EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: req,
	}
}

type RequestSignal struct {
	*Signal `validate:"required"`
	eventloop.EventLoopRequest
	StartFromStepIdx int
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

func (s *Signal) Namespace() string {
	return TemporalNamespace
}

func (s *Signal) Name() string {
	return string(s.Type)
}

func (s *Signal) Stop() bool {
	switch s.Type {
	default:
	}

	return false
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

	ap := app.AppBranch{}
	res := db.WithContext(ctx).
		Preload("Org").
		First(&ap, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &ap.Org, nil
}
