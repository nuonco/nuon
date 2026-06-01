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

	OperationCreated             SignalType = "created"
	OperationPromotion           SignalType = "promotion"
	OperationRestart             SignalType = "restart"
	OperationTerminateEventLoops SignalType = "terminate_event_loops"
	OperationSeed                SignalType = "seed"
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

type RequestSignal struct {
	*Signal
	EventLoopRequest
}

func NewRequestSignal(ev EventLoopRequest, signal *Signal) RequestSignal {
	return RequestSignal{
		Signal:           signal,
		EventLoopRequest: ev,
	}
}

type Signal struct {
	Type SignalType `validate:"required"`

	// Only added when a promotion goes out
	Tag string `json:"tag"`

	CtxPayload      *propagator.Payload `json:"ctx_payload"`
	SignalListeners []SignalListener    `json:"signal_listeners"`
	CGroup          string              `json:"cgroup"`
}

func (s *Signal) WorkflowName() string        { return "EventLoop" }
func (s *Signal) WorkflowID(id string) string { return "event-loop-" + id }
func (s *Signal) Namespace() string           { return TemporalNamespace }
func (s *Signal) Name() string                { return s.Type }
func (s *Signal) SignalType() SignalType      { return s.Type }
func (s *Signal) ConcurrencyGroup() string    { return s.CGroup }
func (s *Signal) Listeners() []SignalListener { return s.SignalListeners }

func (s *Signal) Validate(v *validator.Validate) error {
	if err := v.Struct(s); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}

func (s *Signal) Restart() bool {
	return s.Type == OperationRestart
}

func (s *Signal) Stop() bool {
	return false
}

func (s *Signal) Start() bool {
	return s.Type == OperationCreated
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
	if s.CtxPayload.LogStream != nil {
		ctx = cctx.SetLogStreamWorkflowContext(ctx, s.CtxPayload.LogStream)
	}
	return ctx
}

func (s *Signal) GetContext(ctx context.Context) context.Context {
	if s.CtxPayload == nil {
		return ctx
	}
	ctx = cctx.SetAccountIDContext(ctx, s.CtxPayload.AccountID)
	ctx = cctx.SetOrgIDContext(ctx, s.CtxPayload.OrgID)
	ctx = cctx.SetTraceIDContext(ctx, s.CtxPayload.TraceID)
	if s.CtxPayload.LogStream != nil {
		ctx = cctx.SetLogStreamContext(ctx, s.CtxPayload.LogStream)
	}
	return ctx
}

// NOTE: the general loop has no concept for an organization.
func (s *Signal) GetOrg(ctx context.Context, id string, db *gorm.DB) (*app.Org, error) {
	return nil, nil
}
