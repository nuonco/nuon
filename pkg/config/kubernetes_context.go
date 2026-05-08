package config

import (
	"github.com/invopop/jsonschema"
)

// KubernetesContextsConfig is the top-level container for named Kubernetes
// context bindings. Each context maps a stable name to a peer component that
// emits cluster connection details as outputs (in the same shape the sandbox
// uses today). Components can opt into a context by name; otherwise they fall
// back to the implicit sandbox default when the sandbox emits cluster outputs.
type KubernetesContextsConfig struct {
	Contexts []*KubernetesContext `mapstructure:"kubernetes_context,omitempty" toml:"kubernetes_context,omitempty"`
}

func (a KubernetesContextsConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("kubernetes_context").Short("list of kubernetes contexts").
		Long("Array of named kubernetes context bindings. Each context references a peer component that produces cluster connection details. Components opt into a context via the top-level kubernetes_context field")
}

func (a *KubernetesContextsConfig) parse() error {
	return nil
}

func (a *KubernetesContextsConfig) Validate() error {
	seen := make(map[string]struct{}, len(a.Contexts))
	for _, c := range a.Contexts {
		if err := c.Validate(); err != nil {
			return err
		}
		if _, dup := seen[c.Name]; dup {
			return ErrConfig{
				Description: "duplicate kubernetes_context name: " + c.Name,
			}
		}
		seen[c.Name] = struct{}{}
	}
	return nil
}

// KubernetesContext is a named binding to a peer component that emits cluster
// connection details as outputs. The peer component must be a terraform_module
// or pulumi component, and must expose a `cluster` output object matching the
// per-cloud shape the sandbox uses (see resolveKubernetesContext for the exact
// per-cloud field contract).
//
// Static / external clusters are intentionally not modeled here — wrap them in
// a thin terraform_module component that emits the same cluster outputs.
type KubernetesContext struct {
	Name      string `mapstructure:"name" toml:"name" jsonschema:"required"`
	Component string `mapstructure:"component" toml:"component" jsonschema:"required"`
}

func (a KubernetesContext) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("name").Short("context name").Required().
		Long("Unique name for this kubernetes context within the app. Components reference this name via the top-level kubernetes_context field").
		Example("data-cluster").
		Example("shared-prod").
		Field("component").Short("source peer component").Required().
		Long("Name of the peer component that produces cluster connection details. Must be a terraform_module or pulumi component, and must expose a `cluster` output object").
		Example("data-eks").
		Example("shared-aks")
}

func (a *KubernetesContext) Validate() error {
	if a.Name == "" {
		return ErrConfig{Description: "kubernetes_context.name is required"}
	}
	if a.Component == "" {
		return ErrConfig{Description: "kubernetes_context.component is required"}
	}
	return nil
}
