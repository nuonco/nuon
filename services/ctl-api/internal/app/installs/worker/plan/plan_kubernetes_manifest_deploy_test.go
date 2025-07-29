package plan

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// dummy test just to validate outputs
func TestPlanner_CreateKubernetesManifestDeployPlanSandboxMode(t *testing.T) {
	planner := &Planner{}
	t.Run("successfully creates KubernetesSandboxMode", func(t *testing.T) {
		plan, err := planner.createKubernetesManifestDeployPlanSandboxMode(nil)
		assert.NoError(t, err)
		assert.NotNil(t, plan)
	})
}
