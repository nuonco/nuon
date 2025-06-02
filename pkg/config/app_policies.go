package config

import (
	"github.com/invopop/jsonschema"
)

type PoliciesConfig struct {
	Policies []AppPolicy `mapstructure:"policy,omitempty"`
}

func (a PoliciesConfig) JSONSchemaExtend(schema *jsonschema.Schema) {
	addDescription(schema, "policy", "List of policies to enforce.")
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
	addDescription(schema, "type", "Policy type which controls how it is enforced.")
	addDescription(schema, "contents", "The policy contents. Supports any reference via https://github.com/hashicorp/go-getter.")
}

func (a *AppPolicy) parse() error {
	return nil
}
