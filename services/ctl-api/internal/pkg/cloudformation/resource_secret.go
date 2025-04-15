package cloudformation

import (
	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/secretsmanager"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (a *Templates) getSecretsParamLabels(inp *TemplateInput) map[string]any {
	paramLabels := make(map[string]any, 0)
	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		paramLabels[secret.CloudFormationParamName] = secret.DisplayName
	}

	return paramLabels
}

func (a *Templates) getSecretsParameters(inp *TemplateInput) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		params[secret.CloudFormationParamName] = cloudformation.Parameter{
			Type:        "String",
			Description: generics.ToPtr(secret.Description),
			NoEcho:      generics.ToPtr(true),
		}
	}

	return params
}

func (a *Templates) getSecretsResources(inp *TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	rsrcs := make(map[string]cloudformation.Resource, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		rsrcs[secret.CloudFormationStackName] = &secretsmanager.Secret{
			Name:        generics.ToPtr(secret.Name),
			Description: generics.ToPtr(secret.Description),
			Tags:        t.apply(nil, ""),
		}
	}

	return rsrcs
}
