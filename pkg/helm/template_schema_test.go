package helm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"helm.sh/helm/v4/pkg/action"
	"helm.sh/helm/v4/pkg/chart/common"
	chart "helm.sh/helm/v4/pkg/chart/v2"
	rcommon "helm.sh/helm/v4/pkg/release/common"
	release "helm.sh/helm/v4/pkg/release/v1"
	"k8s.io/client-go/rest"
)

func policySchemaChart(schema string) *chart.Chart {
	ch := &chart.Chart{
		Metadata:  &chart.Metadata{APIVersion: "v2", Name: "acme", Version: "1.0.0"},
		Templates: []*common.File{{Name: "templates/config.yaml", Data: []byte("apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: {{ .Values.configName }}\n")}},
	}
	if schema != "" {
		ch.Schema = []byte(schema)
	}
	return ch
}

func externalPolicySchema(t *testing.T) (string, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "external.schema.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"type":"object","properties":{"configName":{"const":"external-allow"}}}`), 0600))
	return path, (&url.URL{Scheme: "file", Path: path}).String()
}

func TestTemplateChartRejectsExternalSchemaReads(t *testing.T) {
	path, address := externalPolicySchema(t)
	for _, ref := range []string{address, filepath.ToSlash(path)} {
		t.Run(ref, func(t *testing.T) {
			schema := fmt.Sprintf(`{"$ref":%q}`, ref)
			_, err := TemplateChart(policySchemaChart(schema), map[string]interface{}{"configName": "external-allow"})
			require.ErrorContains(t, err, "external schema resource is not allowed")
		})
	}
}

func TestTemplateChartSelfContainedSchemaValidation(t *testing.T) {
	for name, schema := range map[string]string{
		"fragment":            `{"$schema":"https://json-schema.org/draft/2020-12/schema","$defs":{"name":{"const":"valid"}},"properties":{"configName":{"$ref":"#/$defs/name"}}}`,
		"draft7":              `{"$schema":"http://json-schema.org/draft-07/schema#","definitions":{"name":{"const":"valid"}},"properties":{"configName":{"$ref":"#/definitions/name"}}}`,
		"embedded ID":         `{"$id":"https://example.com/root.json","$defs":{"name":{"$id":"name.json","const":"valid"}},"properties":{"configName":{"$ref":"name.json"}}}`,
		"unresolved URN":      `{"allOf":[{"$ref":"urn:acme:external"},{"properties":{"configName":{"const":"valid"}}}]}`,
		"dynamic reference":   `{"$schema":"https://json-schema.org/draft/2020-12/schema","$id":"https://example.com/node","$dynamicAnchor":"node","type":"object","required":["configName"],"properties":{"configName":{"const":"valid"},"child":{"$dynamicRef":"#node"}}}`,
		"recursive reference": `{"$schema":"https://json-schema.org/draft/2019-09/schema","$recursiveAnchor":true,"type":"object","required":["configName"],"properties":{"configName":{"const":"valid"},"child":{"$recursiveRef":"#"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			values := map[string]interface{}{"configName": "valid", "child": map[string]interface{}{"configName": "valid"}}
			manifest, err := TemplateChart(policySchemaChart(schema), values)
			require.NoError(t, err)
			require.Contains(t, manifest, "name: valid")
			values["configName"] = "invalid"
			_, err = TemplateChart(policySchemaChart(schema), values)
			require.ErrorContains(t, err, "valid")
			values["configName"] = "valid"
			values["child"] = map[string]interface{}{"configName": "invalid"}
			if name == "dynamic reference" || name == "recursive reference" {
				_, err = TemplateChart(policySchemaChart(schema), values)
				require.Error(t, err)
			}
		})
	}
}

func TestTemplateChartRejectsResolvedExternalResources(t *testing.T) {
	path, address := externalPolicySchema(t)
	var requests atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		fmt.Fprintf(w, `{"$ref":%q}`, address)
	}))
	t.Cleanup(server.Close)
	for name, schema := range map[string]string{
		"relative file":              `{"$ref":"relative.schema.json"}`,
		"ID rebased relative file":   fmt.Sprintf(`{"$id":%q,"$ref":"external.schema.json"}`, (&url.URL{Scheme: "file", Path: filepath.Join(filepath.Dir(path), "base.json")}).String()),
		"external metaschema":        fmt.Sprintf(`{"$schema":%q,"type":"object"}`, address),
		"HTTP to file chain":         fmt.Sprintf(`{"$ref":%q}`, server.URL+"/schema.json"),
		"external dynamic reference": fmt.Sprintf(`{"$schema":"https://json-schema.org/draft/2020-12/schema","$dynamicRef":%q}`, address),
		"custom scheme":              `{"$ref":"plugin://external/schema.json"}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := TemplateChart(policySchemaChart(schema), map[string]interface{}{"configName": "external-allow"})
			require.ErrorContains(t, err, "external schema resource is not allowed")
			_, err = RenderManifest(policySchemaChart(schema), map[string]interface{}{"configName": "external-allow"}, "acme", "default")
			require.ErrorContains(t, err, "external schema resource is not allowed")
			_, err = HelmInstallWithLogStreaming(context.Background(), action.NewInstall(&action.Configuration{}), policySchemaChart(schema), nil, nil, nil, zap.NewNop())
			require.ErrorContains(t, err, "external schema resource is not allowed")
			_, err = HelmUpgradeWithLogStreaming(context.Background(), action.NewUpgrade(&action.Configuration{}), "acme", policySchemaChart(schema), nil, nil, nil, zap.NewNop())
			require.ErrorContains(t, err, "external schema resource is not allowed")
		})
	}
	require.Zero(t, requests.Load(), "the first HTTP fetch must be rejected, not merely its nested file reference")
}

func TestTemplateChartActiveDependencySchemas(t *testing.T) {
	_, address := externalPolicySchema(t)
	newChart := func() *chart.Chart {
		root := policySchemaChart("")
		child := policySchemaChart(fmt.Sprintf(`{"$ref":%q}`, address))
		child.Metadata.Name = "child"
		root.AddDependency(child)
		root.Metadata.Dependencies = []*chart.Dependency{{Name: "child", Version: "1.0.0", Condition: "child.enabled"}}
		return root
	}
	values := map[string]interface{}{"configName": "root", "child": map[string]interface{}{"enabled": false}}
	ch := newChart()
	_, err := TemplateChart(ch, values)
	require.NoError(t, err)
	values["child"] = map[string]interface{}{"enabled": true, "configName": "external-allow"}
	_, err = TemplateChart(newChart(), values)
	require.ErrorContains(t, err, "chart child")
	require.ErrorContains(t, err, "external schema resource is not allowed")
}

func TestRenderManifestSchemaValidation(t *testing.T) {
	schema := `{"$defs":{"name":{"const":"valid"}},"properties":{"configName":{"$ref":"#/$defs/name"}}}`
	values := map[string]interface{}{"configName": "valid"}
	manifest, err := RenderManifest(policySchemaChart(schema), values, "acme", "default")
	require.NoError(t, err)
	require.Contains(t, manifest, "name: valid")
	values["configName"] = "invalid"
	_, err = RenderManifest(policySchemaChart(schema), values, "acme", "default")
	require.ErrorContains(t, err, "valid")

	_, address := externalPolicySchema(t)
	ch := policySchemaChart("")
	child := policySchemaChart(fmt.Sprintf(`{"$ref":%q}`, address))
	child.Metadata.Name = "child"
	ch.AddDependency(child)
	ch.Metadata.Dependencies = []*chart.Dependency{{Name: "child", Alias: "alias", Version: "1.0.0", Condition: "alias.enabled"}}
	ch.Values = map[string]interface{}{"alias": map[string]interface{}{"enabled": true}}
	values = map[string]interface{}{"configName": "root", "alias": map[string]interface{}{"enabled": false}}
	require.NoError(t, ValidateSchemaResources(ch, values))
	require.Len(t, ch.Dependencies(), 1)
	require.Equal(t, "child", ch.Dependencies()[0].Name())
	require.Equal(t, "child", ch.Metadata.Dependencies[0].Name)
	require.Equal(t, true, ch.Values["alias"].(map[string]interface{})["enabled"])
	values["alias"] = map[string]interface{}{"enabled": true}
	_, err = RenderManifest(ch, values, "acme", "default")
	require.ErrorContains(t, err, "external schema resource is not allowed")
	values["alias"] = map[string]interface{}{"enabled": false}
	_, err = RenderManifest(ch, values, "acme", "default")
	require.NoError(t, err)
}

func TestTemplateChartAliasAndImportSchemaValues(t *testing.T) {
	t.Run("aliases", func(t *testing.T) {
		newChart := func() *chart.Chart {
			root := policySchemaChart("")
			child := policySchemaChart(`{"properties":{"configName":{"const":"valid"}}}`)
			child.Metadata.Name = "child"
			child.Values = map[string]interface{}{"configName": "default-invalid"}
			root.AddDependency(child)
			root.Metadata.Dependencies = []*chart.Dependency{{Name: "child", Alias: "first", Version: "1.0.0"}, {Name: "child", Alias: "second", Version: "1.0.0"}}
			return root
		}
		values := map[string]interface{}{"configName": "root", "first": map[string]interface{}{"configName": "valid"}, "second": map[string]interface{}{"configName": "valid"}}
		manifest, err := TemplateChart(newChart(), values)
		require.NoError(t, err)
		require.Contains(t, manifest, "charts/first/templates/config.yaml")
		require.Contains(t, manifest, "charts/second/templates/config.yaml")
		values["second"] = map[string]interface{}{"configName": "invalid"}
		_, err = TemplateChart(newChart(), values)
		require.ErrorContains(t, err, "second")
		require.ErrorContains(t, err, "valid")
	})
	t.Run("import values", func(t *testing.T) {
		root := policySchemaChart(`{"properties":{"configName":{"const":"imported"}},"required":["configName"]}`)
		child := policySchemaChart("")
		child.Metadata.Name = "child"
		child.Values = map[string]interface{}{"configName": "child", "exports": map[string]interface{}{"defaults": map[string]interface{}{"configName": "imported"}}}
		root.AddDependency(child)
		root.Metadata.Dependencies = []*chart.Dependency{{Name: "child", Version: "1.0.0", ImportValues: []interface{}{"defaults"}}}
		manifest, err := TemplateChart(root, nil)
		require.NoError(t, err)
		require.Contains(t, manifest, "name: imported")
	})
}

func TestStreamingUpgradePinsCheckedValues(t *testing.T) {
	for name, values := range map[string]map[string]interface{}{
		"chart defaults": nil,
		"explicit":       {"configName": "provided"},
	} {
		t.Run(name, func(t *testing.T) {
			previous := &release.Release{Name: "acme", Namespace: "default", Version: 1,
				Info: &release.Info{Status: rcommon.StatusDeployed}, Config: map[string]interface{}{"configName": "unchecked"}}
			cfg := confirmConfig(t, nil, previous)
			cfg.Capabilities = common.DefaultCapabilities.Copy()
			client := DefaultUpgrade(cfg)
			client.DryRunStrategy = action.DryRunClient
			ch := policySchemaChart(`{"properties":{"configName":{"enum":["chart-default","provided"]}}}`)
			ch.Values = map[string]interface{}{"configName": "chart-default"}
			rel, err := HelmUpgradeWithLogStreaming(context.Background(), client, "acme", ch, values, &rest.Config{}, nil, zap.NewNop())
			require.NoError(t, err)
			want := "chart-default"
			if len(values) > 0 {
				want = "provided"
			}
			require.Contains(t, rel.Manifest, "name: "+want)
			require.Equal(t, values, rel.Config)
		})
	}
}
