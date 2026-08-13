package arm

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

type wiringStack struct {
	name            string
	params          map[string]any
	outputs         []string
	parameters      map[string]string
	managedIdentity bool
}

func armWiringInput(t *testing.T, stacksIn ...wiringStack) *stacks.TemplateInput {
	t.Helper()

	templates := map[string]string{}
	for _, s := range stacksIn {
		params := map[string]any{}
		maps.Copy(params, s.params)
		outputs := map[string]any{}
		for _, name := range s.outputs {
			outputs[name] = map[string]any{"type": "string", "value": "x"}
		}

		resources := []any{}
		if s.managedIdentity {
			resources = append(resources, map[string]any{
				"type":       "Microsoft.ManagedIdentity/userAssignedIdentities",
				"apiVersion": "2023-01-31",
				"name":       s.name + "-identity",
			})
		}

		body, err := json.Marshal(map[string]any{
			"$schema":        "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
			"contentVersion": "1.0.0.0",
			"parameters":     params,
			"resources":      resources,
			"outputs":        outputs,
		})
		require.NoError(t, err)
		templates["/"+s.name+".json"] = string(body)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := templates[r.URL.Path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	inp := minimalTemplateInput()
	for i, s := range stacksIn {
		inp.AppCfg.StackConfig.CustomNestedStacks = append(inp.AppCfg.StackConfig.CustomNestedStacks, config.CustomNestedStack{
			Name:        s.name,
			TemplateURL: server.URL + "/" + s.name + ".json",
			Index:       i,
			Parameters:  s.parameters,
		})
	}

	return inp
}

func stringParam() map[string]any {
	return map[string]any{"type": "string"}
}

func TestGetCustomLinkedDeployments_OutputWiring(t *testing.T) {
	t.Run("two stacks may consume the same vnet output", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "db-subnet", params: map[string]any{"vnetName": stringParam()}},
			wiringStack{name: "dns-zones", params: map[string]any{"vnetName": stringParam()}},
		)

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, resources, 2)

		want := "[reference('vnetDeployment').outputs.vnetName.value]"
		assert.Equal(t, want, armDeploymentParamValue(t, resources[0], "vnetName"))
		assert.Equal(t, want, armDeploymentParamValue(t, resources[1], "vnetName"))
		assert.NotContains(t, hoisted, "vnetName")
	})

	t.Run("earlier stack output wires into a later stack parameter", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "db-subnet", outputs: []string{"dbSubnetId"}},
			wiringStack{name: "apps-postgres", params: map[string]any{"dbSubnetId": stringParam()}},
		)

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t,
			"[reference('DbSubnet').outputs.dbSubnetId.value]",
			armDeploymentParamValue(t, resources[1], "dbSubnetId"),
		)
		assert.NotContains(t, hoisted, "dbSubnetId")
	})

	t.Run("genuine conflict on a hoisted param still errors", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "one", params: map[string]any{"instanceSize": stringParam()}},
			wiringStack{name: "two", params: map[string]any{"instanceSize": stringParam()}},
		)

		_, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `parameter "instanceSize" conflicts with stack "one"`)
	})

	t.Run("a stack cannot wire to its own output", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "solo", params: map[string]any{"zoneId": stringParam()}, outputs: []string{"zoneId"}},
		)

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t, "[parameters('zoneId')]", armDeploymentParamValue(t, resources[0], "zoneId"))
		assert.Contains(t, hoisted, "zoneId")
	})

	t.Run("vnet output beats a custom output of the same name", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "shadow", outputs: []string{"vnetName"}},
			wiringStack{name: "consumer", params: map[string]any{"vnetName": stringParam()}},
		)

		resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t,
			"[reference('vnetDeployment').outputs.vnetName.value]",
			armDeploymentParamValue(t, resources[1], "vnetName"),
		)
	})

	t.Run("earliest custom producer wins", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "first", outputs: []string{"sharedId"}},
			wiringStack{name: "second", outputs: []string{"sharedId"}},
			wiringStack{name: "consumer", params: map[string]any{"sharedId": stringParam()}},
		)

		resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t,
			"[reference('First').outputs.sharedId.value]",
			armDeploymentParamValue(t, resources[2], "sharedId"),
		)
	})

	t.Run("explicit parameter beats wiring", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{
				name:       "consumer",
				params:     map[string]any{"vnetName": stringParam()},
				parameters: map[string]string{"vnetName": "explicit-vnet"},
			},
		)

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t, "explicit-vnet", armDeploymentParamValue(t, resources[0], "vnetName"))
		assert.NotContains(t, hoisted, "vnetName")
	})

	t.Run("a custom output cannot shadow a reserved param", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "shadow", outputs: []string{"location", "nuonInstallID"}},
			wiringStack{name: "consumer", params: map[string]any{
				"location":      stringParam(),
				"nuonInstallID": stringParam(),
			}},
		)

		resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t, "[parameters('location')]", armDeploymentParamValue(t, resources[1], "location"))
		assert.Equal(t, inp.Install.ID, armDeploymentParamValue(t, resources[1], "nuonInstallID"))
	})

	t.Run("documented SharedSubnetID example", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "stack-a", outputs: []string{"SharedSubnetID"}},
			wiringStack{name: "stack-b", params: map[string]any{"SharedSubnetID": stringParam()}},
		)

		resources, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t,
			"[reference('StackA').outputs.SharedSubnetID.value]",
			armDeploymentParamValue(t, resources[1], "SharedSubnetID"),
		)
	})

	t.Run("wiring accumulates across a multi-hop chain", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "hop-zero", outputs: []string{"zeroOut"}},
			wiringStack{name: "hop-one", params: map[string]any{"zeroOut": stringParam()}, outputs: []string{"oneOut"}},
			wiringStack{name: "hop-two", params: map[string]any{"zeroOut": stringParam(), "oneOut": stringParam()}},
		)

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)

		assert.Equal(t, "[reference('HopZero').outputs.zeroOut.value]", armDeploymentParamValue(t, resources[1], "zeroOut"))
		assert.Equal(t, "[reference('HopZero').outputs.zeroOut.value]", armDeploymentParamValue(t, resources[2], "zeroOut"))
		assert.Equal(t, "[reference('HopOne').outputs.oneOut.value]", armDeploymentParamValue(t, resources[2], "oneOut"))
		assert.Empty(t, hoisted)
	})

	t.Run("lovable-enterprise-azure stack set generates", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t,
			wiringStack{name: "storage", outputs: []string{"storageAccountName"}},
			wiringStack{
				name:    "db-subnet",
				params:  map[string]any{"vnetName": stringParam()},
				outputs: []string{"dbSubnetId", "privateDnsZoneId"},
			},
			wiringStack{
				name:       "dns_zones",
				params:     map[string]any{"vnetName": stringParam(), "rootDomain": stringParam()},
				outputs:    []string{"publicZoneId", "internalZoneId"},
				parameters: map[string]string{"rootDomain": "example.installs.lovable.dev"},
			},
			wiringStack{name: "apps_postgres", params: map[string]any{
				"dbSubnetId": stringParam(), "privateDnsZoneId": stringParam(),
			}},
			wiringStack{name: "sandbox_proxy_postgres", params: map[string]any{
				"dbSubnetId": stringParam(), "privateDnsZoneId": stringParam(),
			}},
			wiringStack{name: "bauleiter"},
		)

		resources, hoisted, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, resources, 6)

		for _, name := range []string{"vnetName", "dbSubnetId", "privateDnsZoneId", "rootDomain"} {
			assert.NotContains(t, hoisted, name)
		}
		for _, idx := range []int{3, 4} {
			assert.Equal(t,
				"[reference('DbSubnet').outputs.dbSubnetId.value]",
				armDeploymentParamValue(t, resources[idx], "dbSubnetId"),
				fmt.Sprintf("stack %d", idx),
			)
			assert.Equal(t,
				"[reference('DbSubnet').outputs.privateDnsZoneId.value]",
				armDeploymentParamValue(t, resources[idx], "privateDnsZoneId"),
				fmt.Sprintf("stack %d", idx),
			)
		}
	})
}

func TestGetCustomLinkedDeployments_PrincipalIDOutput(t *testing.T) {
	t.Run("exact identityPrincipalId resolves", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t, wiringStack{
			name:            "bauleiter",
			managedIdentity: true,
			outputs:         []string{"identityPrincipalId"},
		})

		_, _, identities, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, identities, 1)
		assert.Equal(t, "identityPrincipalId", identities[0].PrincipalIDOutput)
	})

	t.Run("prefixed output resolves", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t, wiringStack{
			name:            "bauleiter",
			managedIdentity: true,
			outputs: []string{
				"bauleiterIdentityClientId", "bauleiterIdentityId",
				"bauleiterIdentityName", "bauleiterIdentityPrincipalId",
			},
		})

		_, _, identities, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		require.Len(t, identities, 1)
		assert.Equal(t, "bauleiterIdentityPrincipalId", identities[0].PrincipalIDOutput)

		role := tmpl.getCustomDeploymentRoleAssignment(identities[0])
		params := role["properties"].(map[string]any)["parameters"].(map[string]any)
		assert.Equal(t,
			"[reference('Bauleiter').outputs.bauleiterIdentityPrincipalId.value]",
			params["principalID"].(map[string]any)["value"],
		)
	})

	t.Run("exact match wins over a prefixed one", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t, wiringStack{
			name:            "both",
			managedIdentity: true,
			outputs:         []string{"appIdentityPrincipalId", "identityPrincipalId"},
		})

		_, _, identities, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		assert.Equal(t, "identityPrincipalId", identities[0].PrincipalIDOutput)
	})

	t.Run("managed identity without the output fails at generation", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t, wiringStack{
			name:            "bauleiter",
			managedIdentity: true,
			outputs:         []string{"identityId", "identityName"},
		})

		_, _, _, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "declares a managed identity but no output named")
		assert.Contains(t, err.Error(), "identityId")
	})

	t.Run("stack without a managed identity needs no such output", func(t *testing.T) {
		tmpl := &Templates{cfg: &internal.Config{}}
		inp := armWiringInput(t, wiringStack{name: "storage", outputs: []string{"blobEndpoint"}})

		_, _, identities, _, err := tmpl.getCustomLinkedDeployments(inp)
		require.NoError(t, err)
		assert.Empty(t, identities)
	})
}
