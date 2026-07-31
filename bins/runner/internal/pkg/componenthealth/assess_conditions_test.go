package componenthealth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func cr(kind string, conds ...map[string]any) *unstructured.Unstructured {
	obj := map[string]any{
		"apiVersion": "example.com/v1",
		"kind":       kind,
		"metadata":   map[string]any{"name": "x", "namespace": "n"},
	}
	if conds != nil {
		list := make([]any, 0, len(conds))
		for _, c := range conds {
			list = append(list, c)
		}
		obj["status"] = map[string]any{"conditions": list}
	}
	return &unstructured.Unstructured{Object: obj}
}

// A kind the vendored library does not know must not be reported unknown
// forever; unknown means "could not look", which would be a lie here.
func TestAssessResourceFallsBackToConditions(t *testing.T) {
	cases := []struct {
		name       string
		obj        *unstructured.Unstructured
		wantHealth string
		wantMsg    string
	}{
		{"ready true is healthy",
			cr("NodePool", map[string]any{"type": "Ready", "status": "True"}), healthHealthy, ""},
		{"ready false is degraded with its message",
			cr("NodePool", map[string]any{"type": "Ready", "status": "False", "message": "no capacity"}), healthDegraded, "no capacity"},
		{"ready unknown is progressing",
			cr("NodePool", map[string]any{"type": "Ready", "status": "Unknown"}), healthProgressing, ""},
		{"available counts as ready-style",
			cr("ClickHouseInstallation", map[string]any{"type": "Available", "status": "True"}), healthHealthy, ""},
		{"established counts for CRDs",
			cr("CustomResourceDefinition", map[string]any{"type": "Established", "status": "True"}), healthHealthy, ""},
		{"reason is used when message is absent",
			cr("NodePool", map[string]any{"type": "Ready", "status": "False", "reason": "Unschedulable"}), healthDegraded, "Unschedulable"},

		// No signal is not the same as no information.
		{"no conditions is not-applicable", cr("EC2NodeClass"), healthNotApplicable, ""},
		{"unrelated conditions only", cr("StorageClass",
			map[string]any{"type": "SomethingElse", "status": "True"}), healthNotApplicable, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			health, msg, native := assessResource(tc.obj)
			assert.Equal(t, tc.wantHealth, health)
			assert.Equal(t, tc.wantMsg, msg)
			if tc.wantHealth != healthNotApplicable {
				assert.NotEmpty(t, native, "an assessed resource should carry a native status")
			}
		})
	}
}

// Ready wins over other ready-style types when both are present.
func TestAssessConditionsPrefersReady(t *testing.T) {
	health, _, _ := assessResource(cr("Thing",
		map[string]any{"type": "Available", "status": "False", "message": "stale"},
		map[string]any{"type": "Ready", "status": "True"},
	))
	assert.Equal(t, healthHealthy, health)
}
