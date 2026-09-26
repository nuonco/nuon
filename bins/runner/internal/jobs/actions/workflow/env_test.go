package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// Run env vars are per invocation: a runbook step passing POOLS to a named
// action, or an MCP/ad-hoc run. They used to be applied before the step's own
// configured env vars, so a step that declared the same key silently discarded
// the caller's value and every runbook step got the static config instead.
func TestActionStepEnvRunEnvVarsBeatStepConfig(t *testing.T) {
	h := &handler{
		state: &handlerState{
			plan: &plantypes.ActionWorkflowRunPlan{},
			run:  &models.AppInstallActionWorkflowRun{RunEnvVars: map[string]string{"POOLS": "clickhouse-keeper"}},
		},
	}

	env := h.actionStepEnv(nil, map[string]string{"POOLS": "clickhouse-keeper,clickhouse-installation"})

	assert.Equal(t, "clickhouse-keeper", env["POOLS"])
}

func TestActionStepEnvLayerPrecedence(t *testing.T) {
	h := &handler{
		state: &handlerState{
			plan: &plantypes.ActionWorkflowRunPlan{
				BuiltinEnvVars:  map[string]string{"LAYER": "plan-builtin", "PLAN_ONLY": "yes"},
				OverrideEnvVars: map[string]string{"LAYER": "plan-override"},
			},
			run: &models.AppInstallActionWorkflowRun{
				RunEnvVars: map[string]string{"LAYER": "run", "RUN_ONLY": "yes"},
			},
		},
	}

	env := h.actionStepEnv(
		map[string]string{"LAYER": "built-in", "BUILT_IN_ONLY": "yes"},
		map[string]string{"LAYER": "step-config", "STEP_ONLY": "yes"},
	)

	// plan overrides win outright.
	assert.Equal(t, "plan-override", env["LAYER"])

	// every layer still contributes its own keys.
	assert.Equal(t, "yes", env["PLAN_ONLY"])
	assert.Equal(t, "yes", env["BUILT_IN_ONLY"])
	assert.Equal(t, "yes", env["STEP_ONLY"])
	assert.Equal(t, "yes", env["RUN_ONLY"])
}

// The layers are shared plan/run state, so building the env must not write
// through to them.
func TestActionStepEnvDoesNotMutateSources(t *testing.T) {
	builtin := map[string]string{"LAYER": "plan-builtin"}
	runEnv := map[string]string{"OTHER": "run"}
	stepEnv := map[string]string{"LAYER": "step-config"}

	h := &handler{
		state: &handlerState{
			plan: &plantypes.ActionWorkflowRunPlan{BuiltinEnvVars: builtin},
			run:  &models.AppInstallActionWorkflowRun{RunEnvVars: runEnv},
		},
	}

	h.actionStepEnv(nil, stepEnv)

	assert.Equal(t, map[string]string{"LAYER": "plan-builtin"}, builtin)
	assert.Equal(t, map[string]string{"OTHER": "run"}, runEnv)
	assert.Equal(t, map[string]string{"LAYER": "step-config"}, stepEnv)
}
