package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/v2/updateinstallstackoutputs"
)

// WorkflowStepOptions is a functional option for configuring WorkflowStep
type WorkflowStepOptions func(*app.WorkflowStep)

func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}

func WithGroupIdx(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.GroupIdx = n
	}
}

func WithExecutionType(executionType app.WorkflowStepExecutionType) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.ExecutionType = executionType
	}
}

// signalStepMetadata holds the computed step metadata for a given signal type
type signalStepMetadata struct {
	targetType    string
	executionType app.WorkflowStepExecutionType
	retryable     bool
}

// getSignalStepMetadata maps v2 signal types to step metadata (target type, execution type, retryable).
func getSignalStepMetadata(sigType signal.SignalType, planOnly bool) signalStepMetadata {
	meta := signalStepMetadata{
		executionType: app.WorkflowStepExecutionTypeSystem,
		retryable:     true,
	}

	switch sigType {
	case generateinstallstackversion.SignalType, awaitinstallstackversionrun.SignalType, updateinstallstackoutputs.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallStackVersions)
		meta.retryable = false
	case awaitrunnerhealthy.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeRunners)
		meta.retryable = false
	case componentdeployapplyplan.SignalType, componentdeploysyncandplan.SignalType, componentsyncimage.SignalType,
		componentteardownsyncandplan.SignalType, componentteardownapplyplan.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	case provisionsandboxplan.SignalType, provisionsandboxapplyplan.SignalType,
		deprovisionsandboxplan.SignalType, deprovisionsandboxapplyplan.SignalType,
		reprovisionsandboxplan.SignalType, reprovisionsandboxapplyplan.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallSandboxRuns)
	case executeactionworkflow.SignalType, actionworkflowrun.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns)
	case generatestate.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallStates)
	}

	// User execution type signals
	if sigType == awaitinstallstackversionrun.SignalType {
		meta.executionType = app.WorkflowStepExecutionTypeUser
	}

	// Approval execution type signals
	switch sigType {
	case provisionsandboxplan.SignalType, deprovisionsandboxplan.SignalType, reprovisionsandboxplan.SignalType,
		componentdeploysyncandplan.SignalType, componentteardownsyncandplan.SignalType:
		meta.executionType = app.WorkflowStepExecutionTypeApproval
	}

	// Plan-only skip signals
	if planOnly {
		switch sigType {
		case provisionsandboxapplyplan.SignalType, deprovisionsandboxapplyplan.SignalType, reprovisionsandboxapplyplan.SignalType,
			componentdeployapplyplan.SignalType, componentteardownapplyplan.SignalType:
			meta.executionType = app.WorkflowStepExecutionTypeSkipped
		}
	}

	// Generate state is always hidden
	if sigType == generatestate.SignalType {
		meta.executionType = app.WorkflowStepExecutionTypeHidden
	}

	return meta
}

// installSignalStep creates a WorkflowStep from a v2 queue signal
func installSignalStep(ctx workflow.Context, installID, name string, metadata pgtype.Hstore, sig signal.Signal, planOnly bool, opts ...WorkflowStepOptions) (*app.WorkflowStep, error) {
	if sig == nil {
		step := &app.WorkflowStep{
			Name:          name,
			ExecutionType: app.WorkflowStepExecutionTypeSkipped,
			Status:        app.NewCompositeTemporalStatus(ctx, app.StatusPending),
			Metadata:      metadata,
		}
		for _, o := range opts {
			o(step)
		}
		return step, nil
	}

	meta := getSignalStepMetadata(sig.Type(), planOnly)

	step := &app.WorkflowStep{
		Name:           name,
		ExecutionType:  meta.executionType,
		StepTargetType: meta.targetType,
		OwnerID:        installID,
		OwnerType:      "installs",
		Status:         app.NewCompositeTemporalStatus(ctx, app.StatusPending),
		Metadata:       metadata,
		QueueSignal: &signaldb.SignalData{
			Signal: sig,
		},
		Retryable: meta.retryable,
		Skippable: true,
	}

	step.Timeout = signal.DeriveTimeout(sig)

	if mar, ok := sig.(signal.SignalWithMaxAutoRetries); ok {
		maxAutoRetries := mar.MaxAutoRetries(ctx)
		if step.Status.Metadata == nil {
			step.Status.Metadata = make(map[string]any)
		}
		step.Status.Metadata["max_auto_retries"] = maxAutoRetries
	}

	for _, o := range opts {
		o(step)
	}

	return step, nil
}
