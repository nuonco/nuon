// Package signals contains legacy signal types preserved for compilation compatibility.
package signals

import (
	"context"
	"fmt"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx/propagator"
	"go.temporal.io/sdk/workflow"
)

type SignalType = string

const TemporalNamespace string = "components"

const (
	OperationCreated             SignalType = "created"
	OperationRestart             SignalType = "restart"
	OperationBuild               SignalType = "build"
	OperationQueueBuild          SignalType = "queue_build"
	OperationProvision           SignalType = "provision"
	OperationDelete              SignalType = "delete"
	OperationPollDependencies    SignalType = "poll_dependencies"
	OperationConfigCreated       SignalType = "config_created"
	OperationUpdateComponentType SignalType = "update_component_type"
)

type EventLoopRequest struct {
	ID                 string
	SandboxMode        bool
	Version            string
	RestartCount       int
	VersionChangeCount int
}

type SignalListener struct {
	WorkflowID string
	SignalName string
}

type Signal struct {
	Type          SignalType        `validate:"required"`
	BuildID       string            `validate:"required_if=Operation build"`
	ComponentType app.ComponentType `validate:"required_if=Operation update_component_type"`

	CtxPayload      *propagator.Payload `json:"ctx_payload"`
	SignalListeners []SignalListener    `json:"signal_listeners"`
	CGroup          string              `json:"cgroup"`
}

type RequestSignal struct {
	*Signal
	EventLoopRequest
}

func NewRequestSignal(req EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{Signal: signal, EventLoopRequest: req}
}

func (s *Signal) WorkflowName() string        { return "EventLoop" }
func (s *Signal) WorkflowID(id string) string { return "event-loop-" + id }
func (s *Signal) Namespace() string           { return TemporalNamespace }
func (s *Signal) Name() string                { return s.Type }
func (s *Signal) SignalType() SignalType      { return s.Type }
func (s *Signal) ConcurrencyGroup() string    { return s.CGroup }
func (s *Signal) Listeners() []SignalListener { return s.SignalListeners }
func (s *Signal) Stop() bool                  { return s.Type == OperationDelete }
func (s *Signal) Restart() bool               { return s.Type == OperationRestart }
func (s *Signal) Start() bool                 { return s.Type == OperationCreated }

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

func (s *Signal) PropagateContext(ctx cctx.ValueContext) error {
	payload, err := propagator.FetchPayload(ctx)
	if err != nil {
		return err
	}
	s.CtxPayload = payload
	return nil
}

func (s *Signal) GetWorkflowContext(ctx workflow.Context) workflow.Context {
	if s.CtxPayload == nil {
		return ctx
	}
	ctx = cctx.SetAccountIDWorkflowContext(ctx, s.CtxPayload.AccountID)
	ctx = cctx.SetOrgIDWorkflowContext(ctx, s.CtxPayload.OrgID)
	ctx = cctx.SetTraceIDWorkflowContext(ctx, s.CtxPayload.TraceID)
	return ctx
}

func (s *Signal) GetContext(ctx context.Context) context.Context {
	if s.CtxPayload == nil {
		return ctx
	}
	ctx = cctx.SetAccountIDContext(ctx, s.CtxPayload.AccountID)
	ctx = cctx.SetOrgIDContext(ctx, s.CtxPayload.OrgID)
	ctx = cctx.SetTraceIDContext(ctx, s.CtxPayload.TraceID)
	return ctx
}

func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	org, err := cctx.OrgFromContext(ctx)
	if err == nil {
		return org, nil
	}
	comp := app.Component{}
	res := db.WithContext(ctx).Preload("Org").First(&comp, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get component: %w", res.Error)
	}
	return &comp.Org, nil
}
