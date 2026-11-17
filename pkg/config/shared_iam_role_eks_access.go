package config

import (
	"github.com/invopop/jsonschema"
)

const (
	EKSAccessScopeCluster   = "cluster"
	EKSAccessScopeNamespace = "namespace"
)

type AppAWSIAMRoleEKSAccess struct {
	ClusterName      string                         `mapstructure:"cluster_name" toml:"cluster_name" jsonschema:"required" features:"template"`
	KubernetesGroups []string                       `mapstructure:"kubernetes_groups,omitempty" toml:"kubernetes_groups,omitempty"`
	Policies         []AppAWSIAMRoleEKSAccessPolicy `mapstructure:"policies,omitempty" toml:"policies,omitempty"`
}

type AppAWSIAMRoleEKSAccessPolicy struct {
	PolicyARN   string   `mapstructure:"policy_arn" toml:"policy_arn" jsonschema:"required"`
	AccessScope string   `mapstructure:"access_scope,omitempty" toml:"access_scope,omitempty"`
	Namespaces  []string `mapstructure:"namespaces,omitempty" toml:"namespaces,omitempty"`
}

func (a AppAWSIAMRoleEKSAccess) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("cluster_name").Short("EKS cluster the role gets access to").Required().
		Long("Name of the EKS cluster. Supports templating. When it renders empty, the install stack skips the access entry. Sandbox outputs are empty when the first install stack is generated, so a name from sandbox outputs takes effect on the next stack update, and a name from install inputs takes effect on the first apply").
		Example("{{.nuon.inputs.inputs.cluster_name}}").
		Example("{{.nuon.sandbox.outputs.cluster.name}}").
		Field("kubernetes_groups").Short("Kubernetes groups for the access entry").
		Long("Kubernetes RBAC groups the role is mapped to. Use these when cluster RBAC bindings, not access policies, carry the permissions. Groups that start with system: are reserved by EKS and rejected").
		Example("maintenance").
		Field("policies").Short("EKS access policies to associate").
		Long("EKS access policies associated with the role's access entry. Set at least one policy or one Kubernetes group")
}

func (a AppAWSIAMRoleEKSAccessPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("policy_arn").Short("EKS access policy ARN").Required().
		Long("ARN of an EKS cluster access policy").
		Example("arn:aws:eks::aws:cluster-access-policy/AmazonEKSClusterAdminPolicy").
		Example("arn:aws:eks::aws:cluster-access-policy/AmazonEKSViewPolicy").
		Field("access_scope").Short("scope of the policy").
		Long("Whether the policy applies to the whole cluster or to specific namespaces. Defaults to cluster").
		Enum(EKSAccessScopeCluster, EKSAccessScopeNamespace).
		Field("namespaces").Short("namespaces the policy applies to").
		Long("Required when access_scope is namespace, and not allowed otherwise").
		Example("default")
}
