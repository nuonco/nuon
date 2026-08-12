package v2

import (
	"github.com/jackc/pgx/v5/pgtype"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	signaldb "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal/db"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/actionworkflowrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitcomponenthealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitrunnerhealthy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentsyncimage"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownsyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/executeactionworkflow"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generatestate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionrunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/stackrun"
	statepartialgenerate "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/state/statepartialgenerate"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
)

// WorkflowStepOptions is a functional option for configuring WorkflowStep
type WorkflowStepOptions func(*app.WorkflowStep)

func WithSkippable(skippable bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.Skippable = skippable
	}
}

func WithSkipOnFailure(skipOnFailure bool) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.SkipOnFailure = skipOnFailure
	}
}

func WithGroupIdx(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		s.GroupIdx = n
	}
}

// componentMaxAutoRetries looks up the max auto retries for a component from
// the pre-fetched app config, avoiding redundant activity calls.
// componentGateEnabled reports whether the verified-deploy gate should get its
// own workflow step for a component: block_deploy is opt-in per component, so
// step counts only change for components that asked for the gate. Resolution
// goes through the pin+latest-view activity — an app config version only
// carries ccc rows for components changed in that sync, so reading the
// pre-fetched appCfg would silently drop the gate after any unrelated sync.
// Resolution errors fail open (no step): the deploy must not break because
// gate config could not be read, and the signal re-resolves at run time.
func componentGateEnabled(ctx workflow.Context, installID string, componentID string, componentType app.ComponentType) bool {
	if componentType != app.ComponentTypeHelmChart && componentType != app.ComponentTypeKubernetesManifest {
		return false
	}
	// Without the feature the gate signal no-ops, so inserting the step would
	// show a phantom "verify health" step that verifies nothing.
	if enabled, err := activities.AwaitHasFeatureByFeature(ctx, string(app.OrgFeatureComponentHealth)); err != nil || !enabled {
		return false
	}
	ccc, err := activities.AwaitGetCurrentComponentConfig(ctx, &activities.GetCurrentComponentConfigRequest{
		InstallID:   installID,
		ComponentID: componentID,
	})
	if err != nil || ccc == nil {
		return false
	}
	return ccc.HealthCheckEnabled() && ccc.HealthBlocksDeploy()
}

func componentMaxAutoRetries(appCfg *app.AppConfig, componentID string) int {
	for _, ccc := range appCfg.ComponentConfigConnections {
		if ccc.ComponentID == componentID {
			return ccc.GetMaxAutoRetries()
		}
	}
	return 0
}

func WithMaxAutoRetries(n int) WorkflowStepOptions {
	return func(s *app.WorkflowStep) {
		if s.Status.Metadata == nil {
			s.Status.Metadata = make(map[string]any)
		}
		s.Status.Metadata["max_auto_retries"] = n
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
	case generateinstallstackversion.SignalType, awaitinstallstackversionrun.SignalType, stackrun.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallStackVersions)
		meta.retryable = false
	case awaitrunnerhealthy.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeRunners)
		meta.retryable = false
	case awaitcomponenthealthy.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallComponents)
	case componentdeployapplyplan.SignalType, componentdeploysyncandplan.SignalType, componentsyncimage.SignalType,
		componentteardownsyncandplan.SignalType, componentteardownapplyplan.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallDeploys)
	case provisionsandboxplan.SignalType, provisionsandboxapplyplan.SignalType,
		deprovisionsandboxplan.SignalType, deprovisionsandboxapplyplan.SignalType,
		reprovisionsandboxplan.SignalType, reprovisionsandboxapplyplan.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallSandboxRuns)
	case executeactionworkflow.SignalType, actionworkflowrun.SignalType:
		meta.targetType = string(app.WorkflowStepTargetTypeInstallActionWorkflowRuns)
	case generatestate.SignalType, statepartialgenerate.SignalType:
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

	// Plan-only skip signals. The stack pair is here because generating a stack
	// version is itself a write — it creates a stack version row that supersedes
	// the install's active one, mints a service account and runner token, and
	// then parks awaiting a human to apply the stack.
	if planOnly {
		switch sigType {
		case provisionsandboxapplyplan.SignalType, deprovisionsandboxapplyplan.SignalType, reprovisionsandboxapplyplan.SignalType,
			componentdeployapplyplan.SignalType, componentteardownapplyplan.SignalType,
			generateinstallstackversion.SignalType, awaitinstallstackversionrun.SignalType,
			reprovisionrunner.SignalType:
			meta.executionType = app.WorkflowStepExecutionTypeSkipped
		}
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

	// Apply options first so WithMaxAutoRetries can set the metadata before
	// we decide whether to call the (expensive) MaxAutoRetries activity.
	for _, o := range opts {
		o(step)
	}

	// Only call MaxAutoRetries if not already provided via WithMaxAutoRetries.
	if _, alreadySet := step.Status.Metadata["max_auto_retries"]; !alreadySet {
		if mar, ok := sig.(signal.SignalWithMaxAutoRetries); ok {
			maxAutoRetries := mar.MaxAutoRetries(ctx)
			if step.Status.Metadata == nil {
				step.Status.Metadata = make(map[string]any)
			}
			step.Status.Metadata["max_auto_retries"] = maxAutoRetries
		}
	}

	return step, nil
}
