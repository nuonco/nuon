package componenthealth

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	tfjson "github.com/hashicorp/terraform-json"
	"go.uber.org/fx"
	"go.uber.org/zap"

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
}

type TerraformProviderParams struct {
	fx.In

	L *zap.Logger `name:"system"`
}

func NewTerraformProvider(params TerraformProviderParams) *TerraformProvider {
	return &TerraformProvider{
		l:           params.L,
		byComponent: map[string][]*models.ServiceComponentHealthResource{},
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

	p.mu.Lock()
	defer p.mu.Unlock()
	p.byComponent[componentID] = rows
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
