package v2

import (
	"testing"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/awaitinstallstackversionrun"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeployapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentdeploysyncandplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/componentteardownapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/deprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/generateinstallstackversion"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/provisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionrunner"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/reprovisionsandboxapplyplan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func TestPlanOnlySkipsWriteSignals(t *testing.T) {
	skipped := []signal.SignalType{
		provisionsandboxapplyplan.SignalType,
		deprovisionsandboxapplyplan.SignalType,
		reprovisionsandboxapplyplan.SignalType,
		componentdeployapplyplan.SignalType,
		componentteardownapplyplan.SignalType,
		generateinstallstackversion.SignalType,
		awaitinstallstackversionrun.SignalType,
		reprovisionrunner.SignalType,
	}

	for _, sigType := range skipped {
		t.Run(string(sigType), func(t *testing.T) {
			meta := getSignalStepMetadata(sigType, true)
			if meta.executionType != app.WorkflowStepExecutionTypeSkipped {
				t.Fatalf("plan-only: got execution type %q, want %q", meta.executionType, app.WorkflowStepExecutionTypeSkipped)
			}

			if getSignalStepMetadata(sigType, false).executionType == app.WorkflowStepExecutionTypeSkipped {
				t.Fatal("non-plan-only run should still execute this signal")
			}
		})
	}
}

func TestPlanOnlyKeepsPlanSignals(t *testing.T) {
	for _, sigType := range []signal.SignalType{componentdeploysyncandplan.SignalType} {
		if getSignalStepMetadata(sigType, true).executionType == app.WorkflowStepExecutionTypeSkipped {
			t.Fatalf("plan-only run skipped %q, which produces the plan", sigType)
		}
	}
}
