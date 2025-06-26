package cloudformation

import (
	"fmt"

	"github.com/awslabs/goformation/v7/cloudformation"
	"github.com/awslabs/goformation/v7/cloudformation/secretsmanager"

	"github.com/powertoolsdev/mono/pkg/generics"
)

func (a *Templates) getSecretsParamLabels(inp *TemplateInput) map[string]any {
	paramLabels := make(map[string]any, 0)
	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		if secret.AutoGenerate {
			continue
		}

		paramLabels[secret.CloudFormationParamName] = secret.DisplayName
	}

	return paramLabels
}

func (a *Templates) getSecretsParameters(inp *TemplateInput) map[string]cloudformation.Parameter {
	params := make(map[string]cloudformation.Parameter, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		if secret.AutoGenerate {
			continue
		}

		param := cloudformation.Parameter{
			Type:                  "String",
			Description:           generics.ToPtr(secret.Description),
			NoEcho:                generics.ToPtr(true),
			ConstraintDescription: generics.ToPtr("This parameter is required"),
		}
		if secret.Default != "" {
			param.Default = generics.ToPtr(secret.Default)
		}
		if secret.Required {
			param.AllowedPattern = generics.ToPtr[string]("^[^\\s]{1,}$")
		}
		params[secret.CloudFormationParamName] = param
	}

	return params
}

func (a *Templates) getSecretsResources(inp *TemplateInput, t tagBuilder) map[string]cloudformation.Resource {
	// NOTE: secrets names are "{{instal.id}}/{{secret.name}}" to guarantee uniqueness
	rsrcs := make(map[string]cloudformation.Resource, 0)

	for _, secret := range inp.AppCfg.SecretsConfig.Secrets {
		obj := &secretsmanager.Secret{
			Name:        generics.ToPtr(fmt.Sprintf("%s/%s", t.installID, secret.Name)),
			Description: generics.ToPtr(secret.Description),
			Tags:        t.apply(nil, ""),
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
