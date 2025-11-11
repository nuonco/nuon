package config

import (
	"github.com/invopop/jsonschema"
)

type PoliciesConfig struct {
	Policies []AppPolicy `mapstructure:"policy,omitempty"`
}

func (a PoliciesConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("policy").Short("list of policies").
		Long("Array of policy definitions that enforce compliance and security rules across your infrastructure")
}

func (a *PoliciesConfig) parse() error {
	return nil
}

type AppPolicyType string

const (
	AppPolicyTypeKubernetesClusterKyverno        AppPolicyType = "kubernetes_cluster"
	AppPolicyTypeTerraformDeployRunnerJobKyverno AppPolicyType = "runner_job_terraform_deploy"
	AppPolicyTypeHelmDeployRunnerJobKyverno      AppPolicyType = "runner_job_helm_deploy"
	AppPolicyTypeActionWorkflowRunnerJobKyverno  AppPolicyType = "runner_job_action_workflow"
)

type AppPolicy struct {
	Type     AppPolicyType `mapstructure:"type"`
	Contents string        `mapstructure:"contents" features:"get,template"`
}

func (a AppPolicy) JSONSchemaExtend(schema *jsonschema.Schema) {
	NewSchemaBuilder(schema).
		Field("type").Short("policy type").
		Long("Type of policy that determines where and how it is enforced").
		Example("kubernetes_cluster").
		Example("runner_job_terraform_deploy").
		Example("runner_job_helm_deploy").
		Example("runner_job_action_workflow").
		Field("contents").Short("policy document").
		Long("Policy content in the appropriate format for the policy type. Supports Nuon templating and external file sources: HTTP(S) URLs (https://example.com/policy.json), git repositories (git::https://github.com/org/repo//policy.json), file paths (file:///path/to/policy.json), and relative paths (./policy.json)")
}

type AppPolicyList struct {
	Policy []AppPolicy `mapstructure:"policy"`
}

func (a AppPolicyList) JSONSchemaExtend(s *jsonschema.Schema) {
	NewSchemaBuilder(s).
		Field("policy").
		Short("list of policy documents").
		Long("One or more AppPolicy objects").
		MinItems(1)
}
