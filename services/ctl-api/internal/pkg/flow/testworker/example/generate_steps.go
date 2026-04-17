package example

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	qsignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"
)

const FakeGenerateStepsSignalType qsignal.SignalType = "fake-generate-steps"

// StepConfig describes a single step to generate. Each field maps directly
// to the corresponding FakeStepSignal field, giving tests full control over
// every step's interface composition.
type StepConfig struct {
	Behavior               string   `json:"behavior"`
	ExecutionType          string   `json:"execution_type,omitempty"`
	EnableAutoRetry        bool     `json:"enable_auto_retry,omitempty"`
	CustomMaxRetries       int      `json:"custom_max_retries,omitempty"`
	EnableNoOpCheck        bool     `json:"enable_noop_check,omitempty"`
	EnablePolicyEvaluation bool     `json:"enable_policy_evaluation,omitempty"`
	EnableCancel           bool     `json:"enable_cancel,omitempty"`
	CloneStepNames         []string `json:"clone_step_names,omitempty"`
	Retryable              bool     `json:"retryable,omitempty"`
	Skippable              bool     `json:"skippable,omitempty"`
	GroupIdx               int      `json:"group_idx,omitempty"`
}

// FakeGenerateStepsSignal generates workflow steps from a list of StepConfigs.
// It implements the same pattern as the real generate-workflow-steps signal:
// Execute builds the steps, sets done=true, then blocks forever waiting for
// the FetchSteps update handler to return the generated steps.
type FakeGenerateStepsSignal struct {
	WorkflowID string       `json:"workflow_id"`
	Steps      []StepConfig `json:"steps"`

	generatedSteps []*app.WorkflowStep
	done           bool
}

var _ qsignal.Signal = (*FakeGenerateStepsSignal)(nil)
var _ qsignal.SignalWithUpdateHandlers = (*FakeGenerateStepsSignal)(nil)
var _ qsignal.SignalWithFetchSteps = (*FakeGenerateStepsSignal)(nil)

func (s *FakeGenerateStepsSignal) SetWorkflowID(id string) {
	s.WorkflowID = id
}

func (s *FakeGenerateStepsSignal) Type() qsignal.SignalType {
	return FakeGenerateStepsSignalType
}

func (s *FakeGenerateStepsSignal) Validate(_ workflow.Context) error {
	return nil
}

func (s *FakeGenerateStepsSignal) Execute(ctx workflow.Context) error {
	steps := make([]*app.WorkflowStep, len(s.Steps))
	for i, cfg := range s.Steps {
		execType := app.WorkflowStepExecutionTypeSystem
		if cfg.ExecutionType != "" {
			execType = app.WorkflowStepExecutionType(cfg.ExecutionType)
		}

		steps[i] = &app.WorkflowStep{
			Name:          fmt.Sprintf("step-%d-%s", i, cfg.Behavior),
			Idx:           i * 100,
			Status:        app.CompositeStatus{Status: app.StatusPending},
			ExecutionType: execType,
			Retryable:     cfg.Retryable,
			Skippable:     cfg.Skippable,
			GroupIdx:      cfg.GroupIdx,
			QueueSignal: &signaldb.SignalData{
				Signal: &FakeStepSignal{
					Behavior:               cfg.Behavior,
					EnableAutoRetry:        cfg.EnableAutoRetry,
					CustomMaxRetries:       cfg.CustomMaxRetries,
					EnableNoOpCheck:        cfg.EnableNoOpCheck,
					EnablePolicyEvaluation: cfg.EnablePolicyEvaluation,
					CloneStepNames:         cfg.CloneStepNames,
				},
			},
		}
	}

	s.generatedSteps = steps
	s.done = true

	// Block forever — the conductor calls FetchSteps to retrieve the steps,
	// then the handler workflow terminates via its sleep-after timeout.
	return workflow.Await(ctx, func() bool { return false })
}

func (s *FakeGenerateStepsSignal) RegisterUpdateHandlers(ctx workflow.Context) error {
	return workflow.SetUpdateHandlerWithOptions(ctx, "FetchSteps",
		func(ctx workflow.Context) ([]*app.WorkflowStep, error) {
			if err := workflow.Await(ctx, func() bool { return s.done }); err != nil {
				return nil, err
			}
			return s.generatedSteps, nil
		},
		workflow.UpdateHandlerOptions{},
	)
}
