package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/secretsmanager"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

func (a *Templates) secretConditionName(secretParamName string) string {
	return secretParamName + "Provided"
}

func (a *Templates) getSecretsConditions(inp *stacks.TemplateInput) map[string]any {
	conditions := make(map[string]any, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		if secret.AutoGenerate || secret.Required {
			continue
		}

		conditions[a.secretConditionName(secret.CloudFormationParamName)] = cloudformation.Not(
			[]string{cloudformation.Equals(cloudformation.Ref(secret.CloudFormationParamName), "")},
		)
	}

	return conditions
}

func (a *Templates) getSecretsParamLabels(inp *stacks.TemplateInput) map[string]any {
	paramLabels := make(map[string]any, 0)
	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		if secret.AutoGenerate {
			continue
		}

		paramLabels[secret.CloudFormationParamName] = secret.DisplayName
	}

	return paramLabels
}

func (a *Templates) getSecretsParameters(inp *stacks.TemplateInput) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		if secret.AutoGenerate {
			continue
		}

		param := cloudformation.Parameter{
			Type:        "String",
			Description: generics.ToPtr(secret.Description),
			NoEcho:      generics.ToPtr(true),
		}
		if secret.Default != "" {
			param.Default = generics.ToPtr(secret.Default)
		} else if !secret.Required {
			// Optional secrets should not be required by CloudFormation's create-stack UI.
			param.Default = generics.ToPtr("")
		}
		if secret.Required {
			param.ConstraintDescription = generics.ToPtr("This parameter is required")
			param.AllowedPattern = generics.ToPtr[string]("^[^\\s]{1,}$")
		}
		params[secret.CloudFormationParamName] = param
	}

	return params
}

func (a *Templates) getSecretsResources(inp *stacks.TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	// NOTE: secrets names are "{{install.id}}/{{secret.name}}" to guarantee uniqueness
	rsrcs := make(map[string]cloudformation.Resource, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		obj := &secretsmanager.Secret{
			Name:        generics.ToPtr(fmt.Sprintf("%s/%s", t.installID, secret.Name)),
			Description: generics.ToPtr(secret.Description),
			Tags:        t.apply(nil, ""),
		}
		if !secret.AutoGenerate && !secret.Required {
			obj.AWSCloudFormationCondition = a.secretConditionName(secret.CloudFormationParamName)
		}
		if secret.AutoGenerate {
			obj.GenerateSecretString = &secretsmanager.Secret_GenerateSecretString{
				ExcludePunctuation: generics.ToPtr(true),
				PasswordLength:     generics.ToPtr(63),
			}
		} else {
			obj.SecretString = generics.ToPtr(cloudformation.Sub(fmt.Sprintf("${%s}", secret.CloudFormationParamName)))
		}

		rsrcs[secret.CloudFormationStackName] = obj
	}

	return rsrcs
}
