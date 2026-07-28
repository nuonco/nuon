package arm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/types/state"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

const mockARMCustomTemplateJSON = `{
  "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
  "contentVersion": "1.0.0.0",
  "parameters": {
    "rootDomain": { "type": "string" }
  },
  "resources": [],
  "outputs": {
    "zoneName": { "type": "string", "value": "[parameters('rootDomain')]" }
  }
}`

func armCustomStackInput(t *testing.T, serverURL string, parameters map[string]string) *stacks.TemplateInput {
	t.Helper()

	inp := minimalTemplateInput()
	inp.AppCfg.StackConfig.CustomNestedStacks = []config.CustomNestedStack{
		{
			Name:        "route53_zones",
			TemplateURL: serverURL + "/stack.json",
			Index:       0,
			Parameters:  parameters,
		},
	}

	return inp
}

// ARM had no coverage for explicit parameter values. Both forms are exercised: a
// pre-rendered literal (the current path) and the legacy install-input reference
// (the fallback for callers that read config without rendering it).
func TestGetCustomLinkedDeployments_ExplicitParameters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(mockARMCustomTemplateJSON))
	}))
	defer server.Close()

	t.Run("rendered literal passes through", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armCustomStackInput(t, server.URL, map[string]string{
			"rootDomain": "lvbl.team.example.com",
		})

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, resources, 1)

		assert.Equal(t, "lvbl.team.example.com", armDeploymentParamValue(t, resources[0], "rootDomain"))
		assert.NotContains(t, hoisted, "rootDomain")
	})

	t.Run("legacy install input reference still resolves", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armCustomStackInput(t, server.URL, map[string]string{
			"rootDomain": "{{.nuon.install.inputs.root_domain}}",
		})
		val := "legacy.example.com"
		inp.Install.CurrentInstallInputs = &app.InstallInputs{
			Values: pgtype.Hstore{"root_domain": &val},
		}

		resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, resources, 1)

		assert.Equal(t, "legacy.example.com", armDeploymentParamValue(t, resources[0], "rootDomain"))
	})

	t.Run("templated value renders from install state", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armCustomStackInput(t, server.URL, map[string]string{
			"rootDomain": `{{ if .nuon.install.inputs.root_domain }}{{ .nuon.install.inputs.root_domain }}{{ else }}{{ .nuon.install.id }}.example.com{{ end }}`,
		})

		st := state.New()
		st.ID = inp.Install.ID
		st.Inputs = &state.InputsState{Populated: true, Inputs: map[string]string{"root_domain": ""}}
		st.Install = &state.InstallState{Populated: true, ID: st.ID, Inputs: st.Inputs.Inputs}
		data, err := st.AsMap()
		require.NoError(t, err)

		require.NoError(t, config.RenderCustomNestedStackParameters(inp.AppCfg.StackConfig.CustomNestedStacks, data))

		resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, resources, 1)

		assert.Equal(t, inp.Install.ID+".example.com", armDeploymentParamValue(t, resources[0], "rootDomain"))
	})
}

func armDeploymentParamValue(t *testing.T, resource any, name string) any {
	t.Helper()

	res, ok := resource.(map[string]any)
	require.True(t, ok, "resource is not a map")
	props, ok := res["properties"].(map[string]any)
	require.True(t, ok, "properties is not a map")
	params, ok := props["parameters"].(map[string]any)
	require.True(t, ok, "parameters is not a map")
	entry, ok := params[name].(map[string]any)
	require.True(t, ok, "parameter %q not found in %v", name, params)

	return entry["value"]
}
