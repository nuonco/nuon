package componenthealth

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

const (
	providerAWS   = "aws"
	providerGCP   = "gcp"
	providerAzure = "azure"

	componentTypeTerraformModule = "terraform_module"
)

// terraformClouds maps a provider's short name to its cloud. Absent providers
// (null/random/local/tls, kubernetes/helm) are pseudo or already report live.
var terraformClouds = map[string]string{
	"aws":         providerAWS,
	"awscc":       providerAWS,
	"google":      providerGCP,
	"google-beta": providerGCP,
	"googlebeta":  providerGCP,
	"azurerm":     providerAzure,
	"azuread":     providerAzure,
	"azapi":       providerAzure,
	"azurestack":  providerAzure,
}

// TerraformProvider holds the cloud resources of each terraform-module
// component, handed over by the deploy job since the engine can't read state itself.
type TerraformProvider struct {
	l *zap.Logger

	mu          sync.RWMutex
	byComponent map[string][]*models.ServiceComponentHealthResource
	// byRelease maps a helm release this component's terraform manages to the
	// component, so the release's live workloads are attributed to it instead of
	// being dropped as unowned.
	byRelease map[string]string
	// byObject does the same for individual objects a module applies directly
	// (kubectl_manifest), keyed by resourceKey.
	byObject  map[string]string
	kindsSink *ManifestKindsProvider
}

type TerraformProviderParams struct {
	fx.In

	L     *zap.Logger `name:"system"`
	Kinds *ManifestKindsProvider
}

func NewTerraformProvider(params TerraformProviderParams) *TerraformProvider {
	return &TerraformProvider{
		l:           params.L,
		kindsSink:   params.Kinds,
		byComponent: map[string][]*models.ServiceComponentHealthResource{},
		byRelease:   map[string]string{},
		byObject:    map[string]string{},
	}
}

// Set replaces the resources recorded for a component from a freshly applied
// terraform state; a destroy apply (no cloud resources) clears the rows.
func (p *TerraformProvider) Set(componentID string, state *tfjson.State) {
	if componentID == "" {
		return
	}

	rows := terraformResourceRows(state)
	if len(rows) > maxResourcesPerComponent {
		p.l.Warn("truncating terraform resources for component health",
			zap.String("component_id", componentID),
			zap.Int("resources", len(rows)),
			zap.Int("cap", maxResourcesPerComponent),
		)
		rows = rows[:maxResourcesPerComponent]
	}

	releases := terraformHelmReleases(state)
	objects, gvks := terraformManifestObjects(state)

	p.mu.Lock()
	p.byComponent[componentID] = rows

	// Drop this component's previous entries first, so anything removed from the
	// module stops being attributed to it.
	for release, owner := range p.byRelease {
		if owner == componentID {
			delete(p.byRelease, release)
		}
	}
	for key, owner := range p.byObject {
		if owner == componentID {
			delete(p.byObject, key)
		}
	}
	for _, release := range releases {
		p.byRelease[release] = componentID
	}
	for _, key := range objects {
		p.byObject[key] = componentID
	}
	sink := p.kindsSink
	p.mu.Unlock()

	if sink != nil {
		sink.SetKinds(componentID, gvks, objects)
	}
}

// ComponentForObject returns the terraform component that applied an object
// directly, keyed as resourceKey(kind, namespace, name).
func (p *TerraformProvider) ComponentForObject(key string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	componentID, ok := p.byObject[key]
	return componentID, ok
}

// terraformManifestObjects walks state for kubectl_manifest resources and keys
// each applied object. The provider records kind/name/namespace as attributes,
// so the manifest body only needs parsing when they are absent.
//
// This is attribution only: a kind nobody watches still will not be listed, so
// discovering it here cannot make it reportable on its own.
func terraformManifestObjects(state *tfjson.State) ([]string, []schema.GroupVersionKind) {
	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		return nil, nil
	}

	var keys []string
	var gvks []schema.GroupVersionKind
	var walk func(m *tfjson.StateModule)
	walk = func(m *tfjson.StateModule) {
		if m == nil {
			return
		}
		for _, r := range m.Resources {
			if r == nil || r.Mode == tfjson.DataResourceMode {
				continue
			}
			if r.Type != "kubectl_manifest" && r.Type != "kubernetes_manifest" {
				continue
			}
			key, gvk, ok := manifestObject(r.AttributeValues)
			if !ok {
				continue
			}
			keys = append(keys, key)
			if gvk.Kind != "" {
				gvks = append(gvks, gvk)
			}
		}
		for _, child := range m.ChildModules {
			walk(child)
		}
	}
	walk(state.Values.RootModule)

	return keys, gvks
}

func manifestObject(attrs map[string]interface{}) (string, schema.GroupVersionKind, bool) {
	str := func(k string) string { s, _ := attrs[k].(string); return s }

	apiVersion, kind, namespace, name := str("api_version"), str("kind"), str("namespace"), str("name")
	if kind == "" || name == "" || apiVersion == "" {
		bodyAPI, bodyKind, bodyNS, bodyName := manifestFromBody(str("yaml_body"))
		if kind == "" {
			kind = bodyKind
		}
		if name == "" {
			name, namespace = bodyName, bodyNS
		}
		if apiVersion == "" {
			apiVersion = bodyAPI
		}
	}
	if kind == "" || name == "" {
		return "", schema.GroupVersionKind{}, false
	}

	gvk := schema.GroupVersionKind{}
	if apiVersion != "" {
		if gv, err := schema.ParseGroupVersion(apiVersion); err == nil {
			gvk = gv.WithKind(kind)
		}
	}
	return resourceKey(kind, namespace, name), gvk, true
}

func manifestFromBody(body string) (apiVersion, kind, namespace, name string) {
	if body == "" {
		return "", "", "", ""
	}
	var doc struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Metadata   struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"metadata"`
	}
	if err := yaml.Unmarshal([]byte(body), &doc); err != nil {
		return "", "", "", ""
	}
	return doc.APIVersion, doc.Kind, doc.Metadata.Namespace, doc.Metadata.Name
}

// ComponentForRelease returns the terraform component managing a helm release.
func (p *TerraformProvider) ComponentForRelease(release string) (string, bool) {
	if release == "" {
		return "", false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	componentID, ok := p.byRelease[release]
	return componentID, ok
}

// terraformHelmReleases walks state for helm_release resources. A terraform
// module that installs a chart owns real workloads in the cluster; without this
// they carry no nuon labels and no matching chart component, so they were
// dropped as unowned and the component reported only identity rows.
func terraformHelmReleases(state *tfjson.State) []string {
	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		return nil
	}

	var releases []string
	var walk func(m *tfjson.StateModule)
	walk = func(m *tfjson.StateModule) {
		if m == nil {
			return
		}
		for _, r := range m.Resources {
			if r == nil || r.Type != "helm_release" || r.Mode == tfjson.DataResourceMode {
				continue
			}
			if name, ok := r.AttributeValues["name"].(string); ok && name != "" {
				releases = append(releases, name)
			}
		}
		for _, child := range m.ChildModules {
			walk(child)
		}
	}
	walk(state.Values.RootModule)

	return releases
}

// Resources returns the rows recorded for a component, empty when this process
// has not applied the component's state yet.
func (p *TerraformProvider) Resources(componentID string) []*models.ServiceComponentHealthResource {
	p.mu.RLock()
	defer p.mu.RUnlock()

	rows := p.byComponent[componentID]
	if len(rows) == 0 {
		return nil
	}
	out := make([]*models.ServiceComponentHealthResource, len(rows))
	copy(out, rows)
	return out
}

// ComponentIDs returns every component this process has recorded state for.
func (p *TerraformProvider) ComponentIDs() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	out := make([]string, 0, len(p.byComponent))
	for id := range p.byComponent {
		out = append(out, id)
	}
	return out
}

// terraformResourceRows maps state to identity-only health rows, walking child
// modules. Health is unknown — state proves a resource is managed, not that it lives.
func terraformResourceRows(state *tfjson.State) []*models.ServiceComponentHealthResource {
	if state == nil || state.Values == nil || state.Values.RootModule == nil {
		return nil
	}

	var out []*models.ServiceComponentHealthResource
	var walk func(m *tfjson.StateModule)
	walk = func(m *tfjson.StateModule) {
		if m == nil {
			return
		}
		for _, r := range m.Resources {
			if row := terraformResourceRow(r); row != nil {
				out = append(out, row)
			}
		}
		for _, cm := range m.ChildModules {
			walk(cm)
		}
	}
	walk(state.Values.RootModule)

	return out
}

func terraformResourceRow(r *tfjson.StateResource) *models.ServiceComponentHealthResource {
	if r == nil || r.Type == "" || r.Mode == tfjson.DataResourceMode || isDataSourceAddress(r.Address) {
		return nil
	}

	cloud := terraformCloud(r)
	if cloud == "" {
		return nil
	}

	return &models.ServiceComponentHealthResource{
		Provider: cloud,
		Kind:     r.Type,
		Name:     terraformResourceName(r),
		Health:   healthUnknown,
		Details:  terraformResourceDetails(r),
	}
}

// isDataSourceAddress covers state written without a mode, where the address is
// the only place a data source is distinguishable.
func isDataSourceAddress(address string) bool {
	return strings.HasPrefix(address, "data.") || strings.Contains(address, ".data.")
}

func terraformCloud(r *tfjson.StateResource) string {
	if short := providerShortName(r.ProviderName); short != "" {
		return terraformClouds[short]
	}
	if idx := strings.Index(r.Type, "_"); idx > 0 {
		return terraformClouds[r.Type[:idx]]
	}
	return ""
}

// providerShortName reduces a fully qualified provider source address
// ("registry.terraform.io/hashicorp/aws") to its type ("aws").
func providerShortName(providerName string) string {
	if providerName == "" {
		return ""
	}
	if idx := strings.LastIndex(providerName, "/"); idx >= 0 {
		return providerName[idx+1:]
	}
	return providerName
}

// terraformResourceName prefers the cloud-facing name, then id, then the
// configuration name, so count/for_each instances stay distinct.
func terraformResourceName(r *tfjson.StateResource) string {
	for _, key := range []string{"name", "id"} {
		if s, ok := r.AttributeValues[key].(string); ok && s != "" {
			return s
		}
	}
	if r.Index != nil {
		return fmt.Sprintf("%s[%v]", r.Name, r.Index)
	}
	return r.Name
}

func terraformResourceDetails(r *tfjson.StateResource) string {
	tf := map[string]any{}
	if r.Address != "" {
		tf["address"] = r.Address
	}
	if id, ok := r.AttributeValues["id"].(string); ok && id != "" {
		tf["id"] = id
	}
	if region := terraformResourceRegion(r); region != "" {
		tf["region"] = region
	}
	if r.Tainted {
		tf["tainted"] = true
	}
	if len(tf) == 0 {
		return ""
	}

	b, err := json.Marshal(map[string]any{"terraform": tf})
	if err != nil || len(b) > maxDetailsBytes {
		return ""
	}
	return string(b)
}

func terraformResourceRegion(r *tfjson.StateResource) string {
	for _, key := range []string{"region", "location", "zone"} {
		if s, ok := r.AttributeValues[key].(string); ok && s != "" {
			return s
		}
	}
	if arn, ok := r.AttributeValues["arn"].(string); ok {
		if parts := strings.SplitN(arn, ":", 5); len(parts) >= 4 && parts[0] == "arn" {
			return parts[3]
		}
	}
	return ""
}
