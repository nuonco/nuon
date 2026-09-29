package executeworkflowstep

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/pkg/metrics"
	tmetrics "github.com/nuonco/nuon/pkg/temporal/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "execute-workflow-step"

const (
	DenyViolationsKey  = "deny_violations"
	WarnViolationsKey  = "warn_violations"
	PassedPolicyIDsKey = "passed_policy_ids"
)

func stepSignal(step *app.WorkflowStep) signal.Signal {
	if step.QueueSignal != nil && step.QueueSignal.Signal != nil {
		return step.QueueSignal.Signal
	}
	return nil
}

type Signal struct {
	StepID        string `json:"step_id"`
	StepName      string `json:"step_name,omitempty"`
	StepIdx       int    `json:"step_idx"`
	StepGroupID   string `json:"step_group_id,omitempty"`
	GroupIdx      int    `json:"group_idx"`
	GroupRetryIdx int    `json:"group_retry_idx"`
	RetryIndex    int    `json:"retry_index"`
	WorkflowID    string `json:"workflow_id"`

	WorkflowType string `json:"workflow_type,omitempty"`

	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`

	OrgID     string `json:"org_id,omitempty"`
	OrgName   string `json:"org_name,omitempty"`
	OwnerName string `json:"owner_name,omitempty"`

	TargetQueueName string `json:"target_queue_name"`

	TargetQueueID string `json:"target_queue_id,omitempty"`

	DerivedTimeout time.Duration `json:"derived_timeout,omitempty"`

	ResidentFlow bool `json:"resident_flow,omitempty"`

	ResumeApproval bool `json:"resume_approval,omitempty"`

	innerQueueSignalID string
	finished           bool

	retried bool

	approved bool

	canceled bool

	skipped bool

	approvalResponseID   string
	approvalResponseType string

	mw  metrics.Writer
	v   *validator.Validate
	tmw tmetrics.Writer
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
	_ signal.SignalWithUpdateHandlers   = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithTimeout          = (*Signal)(nil)
	_ signal.SignalWithUnboundedTimeout = (*Signal)(nil)
	_ signal.SignalWithParams           = (*Signal)(nil)
)

func (s *Signal) WithParams(params *signal.Params) {
	s.mw = params.MW
	s.v = params.V
}

func (s *Signal) Timeout() time.Duration {
	if s.DerivedTimeout != 0 {
		return s.DerivedTimeout
	}
	return 30 * 24 * time.Hour
}

func (s *Signal) UnboundedTimeout() bool { return s.DerivedTimeout < 0 }

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		OrgID:        s.OrgID,
		OrgName:      s.OrgName,
		StepID:       s.StepID,
		StepName:     s.StepName,
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		OwnerID:      s.OwnerID,
		OwnerType:    s.OwnerType,
		OwnerName:    s.OwnerName,
		Metadata: map[string]any{
			"step_group_id":   s.StepGroupID,
			"step_idx":        s.StepIdx,
			"group_idx":       s.GroupIdx,
			"group_retry_idx": s.GroupRetryIdx,
			"retry_index":     s.RetryIndex,
		},
	}
}

func (s *Signal) workflowTelemetry() cctx.WorkflowTelemetry {
	telemetry := cctx.WorkflowTelemetry{
		OrgID:        s.OrgID,
		OrgName:      s.OrgName,
		WorkflowID:   s.WorkflowID,
		WorkflowType: s.WorkflowType,
		OwnerID:      s.OwnerID,
		OwnerType:    s.OwnerType,
		OwnerName:    s.OwnerName,
	}
	if s.OwnerType == plugins.TableNameOf[app.Install]() {
		telemetry.InstallID = s.OwnerID
		telemetry.InstallName = s.OwnerName
	}
	return telemetry
}

func (s *Signal) RegisterUpdateHandlers(ctx workflow.Context) error {
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "is-retryable",
		s.isRetryableHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "create-step-retry",
		s.createStepRetryHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "approve-plan",
		s.approvePlanHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "cancel-step",
		s.cancelStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	if err := workflow.SetUpdateHandlerWithOptions(ctx, "skip-step",
		s.skipStepHandler, workflow.UpdateHandlerOptions{}); err != nil {
		return err
	}
	return nil
}

var CacheWindow = 5 * time.Second

func (s *Signal) SleepAfter() time.Duration { return CacheWindow }

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.StepID == "" {
		return errors.New("step_id is required")
	}
	if s.WorkflowID == "" {
		return errors.New("workflow_id is required")
	}
	if s.OwnerID == "" {
		return errors.New("owner_id is required")
	}
	if s.OwnerType == "" {
		return errors.New("owner_type is required")
	}
	return nil
}
