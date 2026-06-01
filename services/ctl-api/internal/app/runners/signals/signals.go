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

const TemporalNamespace string = "runners"

const (
	OperationCreated           SignalType = "created"
	OperationRestart           SignalType = "restart"
	OperationDelete            SignalType = "delete"
	OperationForceDelete       SignalType = "force_delete"
	OperationProcessJob        SignalType = "process_job"
	OperationUpdateVersion     SignalType = "update_version"
	OperationHealthcheck       SignalType = "healthcheck"
	OperationOfflineCheck      SignalType = "offline_check"
	OperationGracefulShutdown  SignalType = "graceful_shutdown"
	OperationForceShutdown     SignalType = "force_shutdown"
	OperationFlushOrphanedJobs SignalType = "flush_orphaned_jobs"

	OperationProvisionServiceAccount   SignalType = "provision_service_account"
	OperationReprovisionServiceAccount SignalType = "reprovision_service_account"
	OperationInstallStackVersionRun    SignalType = "install_stack_version_run"

	OperationMngVMShutDown SignalType = "mng-shutdown-vm"
	OperationMngShutDown   SignalType = "mng-shutdown"
	OperationMngUpdate     SignalType = "mng-update"
	OperationMngRestart    SignalType = "mng-restart"
	OperationMngFetchToken SignalType = "mng-fetch-token"

	OperationProvision   SignalType = "provision"
	OperationDeprovision SignalType = "deprovision"
	OperationReprovision SignalType = "reprovision"
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
	Type SignalType `validate:"required"`

	JobID                    string `validate:"required_if=Type job_queued"`
	HealthCheckID            string `validate:"required_if=Type update_version"`
	InstallStackVersionRunID string `validate:"required_if=Type install_stack_version_run"`
	ForceDelete              bool

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
func (s *Signal) Stop() bool                  { return s.Type == OperationDelete || s.Type == OperationForceDelete }
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
	runner := app.Runner{}
	res := db.WithContext(ctx).Preload("Org").First(&runner, "id = ?", id)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get runner: %w", res.Error)
	}
	return &runner.Org, nil
}
