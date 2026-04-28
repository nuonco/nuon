// Package aws renders the install-stacks/aws Terraform module's tfvars file
// for an AWS install. Mirror of internal/pkg/stacks/gcp.
package aws

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// AWSRoleTemplateInput holds the per-role data rendered into the template.
type AWSRoleTemplateInput struct {
	Name              string
	Permissions       string
	ManagedPolicyArns string
}

// AWSSecretTemplateInput holds a non-auto-gen secret definition for the template.
type AWSSecretTemplateInput struct {
	Name        string
	Description string
	Required    bool
	Default     string
}

// AWSTemplateInput extends TemplateInput with pre-marshaled AWS IAM data.
type AWSTemplateInput struct {
	*stacks.TemplateInput

	ControlPlaneAccountIDs string

	ProvisionPermissions   string
	MaintenancePermissions string
	DeprovisionPermissions string

	ProvisionManagedPolicyArns   string
	MaintenanceManagedPolicyArns string
	DeprovisionManagedPolicyArns string

	BreakGlassRoles []AWSRoleTemplateInput
	CustomRoles     []AWSRoleTemplateInput
	InstallInputs   []string

	AutoGenerateSecrets []string
	Secrets             []AWSSecretTemplateInput
}

// Render emits a JSON-wrapped tfvars envelope for the install-stacks/aws module.
//
// v1: custom nested stacks are not supported — if the app config defines any,
// returns an error and the install must stay on the CloudFormation path.
func Render(inputs *stacks.TemplateInput) ([]byte, string, error) {
	if inputs.AppCfg != nil && len(inputs.AppCfg.StackConfig.CustomNestedStacks) > 0 {
		return nil, "", errors.New("install-stacks/aws Terraform module does not yet support custom nested stacks; keep this install on the CloudFormation path")
	}

	t, err := template.New("aws-stack").Parse(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse aws template")
	}

	prov, maint, deprov, provMPAs, maintMPAs, deprovMPAs := extractAWSStandardPermissions(inputs.AppCfg)
	breakGlassRoles := extractAWSRolesFromList(inputs.AppCfg.BreakGlassConfig.Roles)
	customRoles := extractAWSRolesFromList(inputs.AppCfg.PermissionsConfig.CustomRoles)

	var installInputs []string
	if inputs.AppCfg != nil {
		for _, input := range inputs.AppCfg.InputConfig.AppInputs {
			if input.Source == app.AppInputSourceCustomer {
				installInputs = append(installInputs, input.Name)
			}
		}
	}

	var autoGenerateSecrets []string
	var secrets []AWSSecretTemplateInput
	if inputs.AppCfg != nil {
		for _, s := range inputs.AppCfg.SecretsConfig.Secrets {
			if s.AutoGenerate {
				autoGenerateSecrets = append(autoGenerateSecrets, s.Name)
			} else {
				secrets = append(secrets, AWSSecretTemplateInput{
					Name:        s.Name,
					Description: s.Description,
					Required:    s.Required,
					Default:     s.Default,
				})
			}
		}
	}

	cpAccounts := "[]"
	if b, err := json.Marshal([]string{}); err == nil {
		cpAccounts = string(b)
	}

	awsInputs := &AWSTemplateInput{
		TemplateInput:                inputs,
		ControlPlaneAccountIDs:       cpAccounts,
		ProvisionPermissions:         prov,
		MaintenancePermissions:       maint,
		DeprovisionPermissions:       deprov,
		ProvisionManagedPolicyArns:   provMPAs,
		MaintenanceManagedPolicyArns: maintMPAs,
		DeprovisionManagedPolicyArns: deprovMPAs,
		BreakGlassRoles:              breakGlassRoles,
		CustomRoles:                  customRoles,
		InstallInputs:                installInputs,
		AutoGenerateSecrets:          autoGenerateSecrets,
		Secrets:                      secrets,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, awsInputs); err != nil {
		return nil, "", errors.Wrap(err, "unable to execute aws template")
	}

	envelope := map[string]string{"tfvars": buf.String()}
	res, err := json.Marshal(envelope)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal aws tfvars envelope")
	}

	hash := sha256.Sum256(res)
	return res, hex.EncodeToString(hash[:]), nil
}

// extractAWSStandardPermissions reads AWS IAM policy data for the standard
// roles. Inline-policy contents are not translated — only managed-policy
// attachments are surfaced as ARNs. Inline-permission extraction from policy
// JSON contents is a TODO.
func extractAWSStandardPermissions(appCfg *app.AppConfig) (provision, maintenance, deprovision, provMPAs, maintMPAs, deprovMPAs string) {
	provision = "[]"
	maintenance = "[]"
	deprovision = "[]"
	provMPAs = "[]"
	maintMPAs = "[]"
	deprovMPAs = "[]"

	if appCfg == nil {
		return
	}

	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}

		mpas := managedPolicyArnsForRole(role)
		if len(mpas) == 0 {
			continue
		}
		b, err := json.Marshal(mpas)
		if err != nil {
			continue
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provMPAs = string(b)
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintMPAs = string(b)
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovMPAs = string(b)
		}
	}

	return
}

// extractAWSRolesFromList converts a slice of role configs into template-ready
// inputs, filtering to AWS roles only.
func extractAWSRolesFromList(roles []app.AppAWSIAMRoleConfig) []AWSRoleTemplateInput {
	var result []AWSRoleTemplateInput
	for _, role := range roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}
		mpas := managedPolicyArnsForRole(role)
		if len(mpas) == 0 {
			continue
		}
		b, err := json.Marshal(mpas)
		if err != nil {
			continue
		}
		result = append(result, AWSRoleTemplateInput{
			Name:              role.Name,
			Permissions:       "[]",
			ManagedPolicyArns: string(b),
		})
	}
	return result
}

func managedPolicyArnsForRole(role app.AppAWSIAMRoleConfig) []string {
	var out []string
	for _, policy := range role.Policies {
		if policy.ManagedPolicyName == "" {
			continue
		}
		out = append(out, fmt.Sprintf("arn:aws:iam::aws:policy/%s", policy.ManagedPolicyName))
	}
	return out
}
