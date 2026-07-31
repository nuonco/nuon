package componenthealth

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func tfStateWithReleases(names ...string) *tfjson.State {
	res := make([]*tfjson.StateResource, 0, len(names))
	for _, n := range names {
		res = append(res, &tfjson.StateResource{
			Type:            "helm_release",
			Address:         "helm_release." + n,
			AttributeValues: map[string]interface{}{"name": n},
		})
	}
	return &tfjson.State{Values: &tfjson.StateValues{
		RootModule: &tfjson.StateModule{Resources: res},
	}}
}

func TestTerraformHelmReleases(t *testing.T) {
	assert.ElementsMatch(t, []string{"temporal", "datadog"},
		terraformHelmReleases(tfStateWithReleases("temporal", "datadog")))
	assert.Empty(t, terraformHelmReleases(nil))

	// A data source is a read, not something the module owns.
	assert.Empty(t, terraformHelmReleases(&tfjson.State{Values: &tfjson.StateValues{
		RootModule: &tfjson.StateModule{Resources: []*tfjson.StateResource{{
			Type: "helm_release", Mode: tfjson.DataResourceMode,
			AttributeValues: map[string]interface{}{"name": "read-only"},
		}}},
	}}))
}

func TestComponentForRelease(t *testing.T) {
	p := NewTerraformProvider(TerraformProviderParams{L: zap.NewNop()})

	p.Set("cmp-a", tfStateWithReleases("temporal"))
	id, ok := p.ComponentForRelease("temporal")
	assert.True(t, ok)
	assert.Equal(t, "cmp-a", id)

	_, ok = p.ComponentForRelease("")
	assert.False(t, ok)

	// Re-applying without a release must stop attributing it, or a chart removed
	// from the module keeps reporting against the component forever.
	p.Set("cmp-a", tfStateWithReleases())
	_, ok = p.ComponentForRelease("temporal")
	assert.False(t, ok)
}
