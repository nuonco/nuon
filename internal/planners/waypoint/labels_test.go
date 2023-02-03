package waypoint

import (
	"testing"

	"github.com/google/uuid"
	"github.com/powertoolsdev/go-generics"
	planv1 "github.com/powertoolsdev/protos/workflows/generated/types/executors/v1/plan/v1"
	"github.com/stretchr/testify/assert"
)

func TestDefaultLabels(t *testing.T) {
	meta := generics.GetFakeObj[*planv1.Metadata]()
	compName := uuid.NewString()
	phaseName := uuid.NewString()

	labels := DefaultLabels(meta, compName, phaseName)
	expectedKVs := [][2]string{
		{"deployment-id", meta.DeploymentShortId},
		{"app-id", meta.AppShortId},
		{"org-id", meta.OrgShortId},
		{"component-name", compName},
		{"phase-name", phaseName},
		{"install-id", meta.InstallShortId},
	}
	for _, kv := range expectedKVs {
		assert.Equal(t, labels[kv[0]], kv[1])
	}
	assert.Equal(t, len(labels), len(expectedKVs))

	// ensure that install id is not set empty
	meta.InstallShortId = ""
	labels = DefaultLabels(meta, compName, phaseName)
	expectedKVs = [][2]string{
		{"deployment-id", meta.DeploymentShortId},
		{"app-id", meta.AppShortId},
		{"org-id", meta.OrgShortId},
		{"component-name", compName},
		{"phase-name", phaseName},
	}
	for _, kv := range expectedKVs {
		assert.Equal(t, labels[kv[0]], kv[1])
	}
	assert.Equal(t, len(labels), len(expectedKVs))
}
