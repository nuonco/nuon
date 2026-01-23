package config

import (
	"github.com/invopop/jsonschema"
)

type PoliciesConfig struct {
	Policies []AppPolicy `mapstructure:"policy,omitempty" toml:"policy,omitempty"`
}

func (a PoliciesConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("policy").Short("list of policies").
		Long("Array of policy definitions that enforce compliance and security rules across your infrastructure")
}

func (a *PoliciesConfig) parse() error {
	for i := range a.Policies {
		// to maintain backwards compatibility, default engine based on type
		if a.Policies[i].Engine == "" {
			a.Policies[i].Engine = AppPolicyEngineKyverno
		}
	}
	return nil
}

type AppPolicyType string

const (
	// AppPolicyTypeKubernetesCluster applies to kubernetes cluster-level resources (e.g., namespaces, CRDs)
	AppPolicyTypeKubernetesCluster AppPolicyType = "kubernetes_cluster"
	// AppPolicyTypeTerraformModule applies to terraform module components
	AppPolicyTypeTerraformModule AppPolicyType = "terraform_module"
	// AppPolicyTypeHelmChart applies to helm chart components
	AppPolicyTypeHelmChart AppPolicyType = "helm_chart"
	// AppPolicyTypeKubernetesManifest applies to kubernetes manifest components
	AppPolicyTypeKubernetesManifest AppPolicyType = "kubernetes_manifest"
	// AppPolicyTypeDockerBuild applies to docker build components
	AppPolicyTypeDockerBuild AppPolicyType = "docker_build"
	// AppPolicyTypeContainerImage applies to container image components
	AppPolicyTypeContainerImage AppPolicyType = "container_image"
	// AppPolicyTypeSandbox applies to sandbox infrastructure
	AppPolicyTypeSandbox AppPolicyType = "sandbox"
)

// AllAppPolicyTypes contains all valid policy types
var AllAppPolicyTypes = []AppPolicyType{
	AppPolicyTypeKubernetesCluster,
	AppPolicyTypeTerraformModule,
	AppPolicyTypeHelmChart,
	AppPolicyTypeKubernetesManifest,
	AppPolicyTypeDockerBuild,
	AppPolicyTypeContainerImage,
	AppPolicyTypeSandbox,
}

type AppPolicyEngine string

const (
	AppPolicyEngineKyverno AppPolicyEngine = "kyverno"
	AppPolicyEngineOPA     AppPolicyEngine = "opa"
)

// AllAppPolicyEngines contains all valid policy engines
var AllAppPolicyEngines = []AppPolicyEngine{
	AppPolicyEngineKyverno,
	AppPolicyEngineOPA,
}

type AppPolicy struct {
	Type       AppPolicyType   `mapstructure:"type"`
	Engine     AppPolicyEngine `mapstructure:"engine,omitempty"`
	Contents   string          `mapstructure:"contents" features:"get,template"`
	Components []string        `mapstructure:"components,omitempty"`
}

func (a AppPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("policy type").
		Long("Type of policy that determines where and how it is enforced").
		Enum(
			string(AppPolicyTypeKubernetesCluster),
			string(AppPolicyTypeTerraformModule),
			string(AppPolicyTypeHelmChart),
			string(AppPolicyTypeKubernetesManifest),
			string(AppPolicyTypeSandbox),
		).
		Field("engine").Short("policy engine").
		Long("The policy engine used to evaluate the policy. Must be compatible with the policy type.").
		Enum(string(AppPolicyEngineKyverno), string(AppPolicyEngineOPA)).
		Field("contents").Short("policy document").
		Long("Policy content in the appropriate format for the policy type. Supports Nuon templating and external file sources: HTTP(S) URLs (https://example.com/policy.json), git repositories (git::https://github.com/org/repo//policy.json), file paths (file:///path/to/policy.json), and relative paths (./policy.json)").
		Field("components").Short("target components").
		Long("List of component names this policy applies to. Use [\"*\"] to apply to all components of the specified type. If empty, doesn't apply to any component. Ignored when type is 'sandbox'.").
		Example("*").
		Example("rds_cluster")
}

type AppPolicyList struct {
	Policy []AppPolicy `mapstructure:"policy" toml:"policy"`
}

func (a AppPolicyList) JSONSchemaExtend(s *jsonschema.Schema) {
	NewSchemaBuilder(s).
		Field("policy").
		Short("list of policy documents").
		Long("One or more AppPolicy objects").
		MinItems(1)
}
