package componenthealth

import (
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

func tfStateWithManifest(attrs map[string]interface{}) *tfjson.State {
	return &tfjson.State{Values: &tfjson.StateValues{
		RootModule: &tfjson.StateModule{Resources: []*tfjson.StateResource{{
			Type: "kubectl_manifest", Address: "kubectl_manifest.x", AttributeValues: attrs,
		}}},
	}}
}

func TestTerraformManifestObjects(t *testing.T) {
	t.Run("explicit attributes are preferred", func(t *testing.T) {
		keys, _ := terraformManifestObjects(tfStateWithManifest(map[string]interface{}{
			"kind": "ClickHouseInstallation", "namespace": "clickhouse", "name": "ch",
		}))
		assert.Equal(t, []string{resourceKey("ClickHouseInstallation", "clickhouse", "ch")}, keys)
	})

	t.Run("falls back to parsing the manifest body", func(t *testing.T) {
		keys, gvks := terraformManifestObjects(tfStateWithManifest(map[string]interface{}{
			"yaml_body": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: web\n  namespace: apps\n",
		}))
		assert.Equal(t, []string{resourceKey("Deployment", "apps", "web")}, keys)
		assert.Equal(t, []schema.GroupVersionKind{{Group: "apps", Version: "v1", Kind: "Deployment"}}, gvks)
	})

	t.Run("unusable entries are skipped", func(t *testing.T) {
		keys, _ := terraformManifestObjects(tfStateWithManifest(map[string]interface{}{}))
		assert.Empty(t, keys)
		keys, _ = terraformManifestObjects(tfStateWithManifest(map[string]interface{}{"yaml_body": "not: yaml: ["}))
		assert.Empty(t, keys)
	})
}

func TestComponentForObject(t *testing.T) {
	p := NewTerraformProvider(TerraformProviderParams{L: zap.NewNop()})
	key := resourceKey("ClickHouseInstallation", "clickhouse", "ch")

	p.Set("cmp-ch", tfStateWithManifest(map[string]interface{}{
		"kind": "ClickHouseInstallation", "namespace": "clickhouse", "name": "ch",
	}))
	id, ok := p.ComponentForObject(key)
	assert.True(t, ok)
	assert.Equal(t, "cmp-ch", id)

	// Removed from the module: attribution must not linger.
	p.Set("cmp-ch", tfStateWithReleases())
	_, ok = p.ComponentForObject(key)
	assert.False(t, ok)
}

// Terraform kinds must persist from the deploy alone. They used to depend on the
// health engine having booted first and wiring a sink, so a deploy that landed
// earlier — or a runner whose engine never runs — dropped them silently.
func TestTerraformPersistsKindsWithoutEngineBoot(t *testing.T) {
	store := &ClusterProvider{l: zap.NewNop(), sandboxReleases: map[string]struct{}{}}
	kinds := NewManifestKindsProvider(ManifestKindsProviderParams{L: zap.NewNop(), Cluster: store})

	p := NewTerraformProvider(TerraformProviderParams{L: zap.NewNop(), Kinds: kinds})
	p.Set("cmp-tf", tfStateWithManifest(map[string]interface{}{
		"api_version": "cert-manager.io/v1",
		"kind":        "Certificate",
		"namespace":   "whoami",
		"name":        "c",
	}))

	assert.Equal(t, []string{"cmp-tf|cert-manager.io/v1/Certificate"}, store.ComponentKinds())
	assert.Len(t, kinds.DiscoveredGVKs(), 1)
}
