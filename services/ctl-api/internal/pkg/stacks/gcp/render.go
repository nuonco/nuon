package gcp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"text/template"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/stacks"
)

// GCPRoleTemplateInput holds the per-role data rendered into the template.
type GCPRoleTemplateInput struct {
	Name           string
	Permissions    string
	PredefinedRole string
}

// GCPSecretTemplateInput holds a non-auto-gen secret definition for the template.
type GCPSecretTemplateInput struct {
	Name        string
	Description string
	Required    bool
	Default     string
}

// GCPTemplateInput extends TemplateInput with pre-marshaled GCP IAM permission lists.
type GCPTemplateInput struct {
	*stacks.TemplateInput
	ProvisionPermissions   string
	MaintenancePermissions string
	DeprovisionPermissions string

	ProvisionPredefinedRole   string
	MaintenancePredefinedRole string
	DeprovisionPredefinedRole string

	BreakGlassRoles []GCPRoleTemplateInput
	CustomRoles     []GCPRoleTemplateInput
	InstallInputs   []string

	AutoGenerateSecrets []string
	Secrets             []GCPSecretTemplateInput
}

func Render(inputs *stacks.TemplateInput) ([]byte, string, error) {
	t, err := template.New("gcp-stack").Parse(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse gcp template")
	}

	prov, maint, deprov, provPredefined, maintPredefined, deprovPredefined := extractGCPStandardPermissions(inputs.AppCfg)
	breakGlassRoles := extractGCPRolesFromList(inputs.AppCfg.BreakGlassConfig.Roles)
	customRoles := extractGCPRolesFromList(inputs.AppCfg.PermissionsConfig.CustomRoles)

	var installInputs []string
	if inputs.AppCfg != nil {
		for _, input := range inputs.AppCfg.InputConfig.AppInputs {
			if input.Source == app.AppInputSourceCustomer {
				installInputs = append(installInputs, input.Name)
			}
		}
	}

	var autoGenerateSecrets []string
	var secrets []GCPSecretTemplateInput
	if inputs.AppCfg != nil {
		for _, s := range inputs.AppCfg.SecretsConfig.Secrets {
			if s.AutoGenerate {
				autoGenerateSecrets = append(autoGenerateSecrets, s.Name)
			} else {
				secrets = append(secrets, GCPSecretTemplateInput{
					Name:        s.Name,
					Description: s.Description,
					Required:    s.Required,
					Default:     s.Default,
				})
			}
		}
	}

	gcpInputs := &GCPTemplateInput{
		TemplateInput:             inputs,
		ProvisionPermissions:      prov,
		MaintenancePermissions:    maint,
		DeprovisionPermissions:    deprov,
		ProvisionPredefinedRole:   provPredefined,
		MaintenancePredefinedRole: maintPredefined,
		DeprovisionPredefinedRole: deprovPredefined,
		BreakGlassRoles:           breakGlassRoles,
		CustomRoles:               customRoles,
		InstallInputs:             installInputs,
		AutoGenerateSecrets:       autoGenerateSecrets,
		Secrets:                   secrets,
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, gcpInputs)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp template")
	}

	// Wrap tfvars in a JSON envelope so it can be stored in the jsonb column.
	// The raw tfvars text is HCL, not valid JSON.
	envelope := map[string]string{"tfvars": buf.String()}
	res, err := json.Marshal(envelope)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to marshal gcp tfvars envelope")
	}

	hash := sha256.Sum256(res)
	checksum := hex.EncodeToString(hash[:])

	return res, checksum, nil
}

// extractGCPStandardPermissions reads GCP IAM permissions for the standard roles (provision, maintenance, deprovision).
func extractGCPStandardPermissions(appCfg *app.AppConfig) (provision, maintenance, deprovision, provPredefined, maintPredefined, deprovPredefined string) {
	provision = "[]"
	maintenance = "[]"
	deprovision = "[]"

	if appCfg == nil {
		return
	}

	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "gcp" {
			continue
		}

		perms, predefinedRole := extractRolePermissions(role)
		if len(perms) == 0 && predefinedRole == "" {
			continue
		}

		if len(perms) > 0 {
			b, err := json.Marshal(perms)
			if err != nil {
				continue
			}

			switch role.Type {
			case app.AWSIAMRoleTypeRunnerProvision:
				provision = string(b)
			case app.AWSIAMRoleTypeRunnerMaintenance:
				maintenance = string(b)
			case app.AWSIAMRoleTypeRunnerDeprovision:
				deprovision = string(b)
			}
		}

		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provPredefined = predefinedRole
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintPredefined = predefinedRole
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovPredefined = predefinedRole
		}
	}

	return
}

// extractGCPRolesFromList converts a slice of role configs into template-ready inputs,
// filtering to GCP roles only.
func extractGCPRolesFromList(roles []app.AppAWSIAMRoleConfig) []GCPRoleTemplateInput {
	var result []GCPRoleTemplateInput
	for _, role := range roles {
		if role.CloudPlatform != "gcp" {
			continue
		}

		perms, predefinedRole := extractRolePermissions(role)
		if len(perms) == 0 && predefinedRole == "" {
			continue
		}

		permStr := "[]"
		if len(perms) > 0 {
			b, err := json.Marshal(perms)
			if err != nil {
				continue
			}
			permStr = string(b)
		}

		result = append(result, GCPRoleTemplateInput{
			Name:           role.Name,
			Permissions:    permStr,
			PredefinedRole: predefinedRole,
		})
	}

	return result
}

func extractRolePermissions(role app.AppAWSIAMRoleConfig) ([]string, string) {
	var perms []string
	var predefinedRole string
	for _, policy := range role.Policies {
		perms = append(perms, policy.GCPPermissions...)
		if policy.GCPPredefinedRole != "" {
			predefinedRole = policy.GCPPredefinedRole
		}
	}
	return perms, predefinedRole
}
