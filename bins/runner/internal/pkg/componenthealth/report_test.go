package componenthealth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func TestEngineCollectProbes(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer healthy.Close()
	failing := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer failing.Close()

	healthySpec, ok := newProbeSpec("http", healthy.URL+"/healthz")
	require.True(t, ok)
	failingSpec, ok := newProbeSpec("http", failing.URL+"/healthz")
	require.True(t, ok)

	e := &Engine{l: zap.NewNop(), idx: newIndex()}
	e.idx.replace("ins-1", []componentEntry{
		{
			installComponentID: "ic-1",
			componentID:        "cmp-1",
			componentType:      "helm_chart",
			probes:             []probeSpec{healthySpec, failingSpec},
		},
		{installComponentID: "ic-2", componentID: "cmp-2", componentType: "helm_chart"},
	})

	grouped := map[string][]*models.ServiceComponentHealthResource{}
	e.collectProbes(context.Background(), grouped)

	require.Len(t, grouped, 1)
	require.Len(t, grouped["cmp-1"], 2)

	byName := map[string]*models.ServiceComponentHealthResource{}
	for _, row := range grouped["cmp-1"] {
		assert.Equal(t, providerProbe, row.Provider)
		assert.Equal(t, resourceKindHTTPProbe, row.Kind)
		byName[row.Name] = row
	}
	assert.Equal(t, healthHealthy, byName[healthySpec.target].Health)
	assert.Equal(t, healthUnhealthy, byName[failingSpec.target].Health)
	assert.Contains(t, byName[failingSpec.target].Message, "503")
}

func TestEngineCollectTerraform(t *testing.T) {
	state := &tfjson.State{Values: &tfjson.StateValues{RootModule: &tfjson.StateModule{
		Resources: []*tfjson.StateResource{
			managed("aws_db_instance.main", "aws_db_instance", "main",
				"registry.terraform.io/hashicorp/aws", map[string]any{"id": "db-1"}),
		},
	}}}

	e := &Engine{
		l:         zap.NewNop(),
		idx:       newIndex(),
		terraform: NewTerraformProvider(TerraformProviderParams{L: zap.NewNop()}),
	}
	e.idx.replace("ins-1", []componentEntry{
		{installComponentID: "ic-1", componentID: "cmp-1", componentType: componentTypeTerraformModule},
		{installComponentID: "ic-2", componentID: "cmp-2", componentType: componentTypeTerraformModule},
	})
	e.terraform.Set("cmp-1", state)
	// a component that has since been removed from the install
	e.terraform.Set("cmp-removed", state)

	grouped := map[string][]*models.ServiceComponentHealthResource{}
	e.collectTerraform(grouped)

	require.Len(t, grouped, 1)
	require.Len(t, grouped["cmp-1"], 1)
	assert.Equal(t, providerAWS, grouped["cmp-1"][0].Provider)
	assert.Equal(t, "aws_db_instance", grouped["cmp-1"][0].Kind)
	assert.Equal(t, healthUnknown, grouped["cmp-1"][0].Health)
}
