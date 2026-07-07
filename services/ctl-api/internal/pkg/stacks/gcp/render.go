package gcp

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
	// Value is what the secret's `value` renders to in tfvars. For the normal
	// tfvars it's the secret's Default; for the Spacelift blueprint it's a
	// `${{ inputs.secret_<name> }}` CEL reference.
	Value string
}

// GCPInstallInputTemplateInput holds a customer install input for the template.
// Value is the input's Default in the normal tfvars and a
// `${{ inputs.input_<name> }}` CEL reference in the Spacelift blueprint.
type GCPInstallInputTemplateInput struct {
	Name    string
	Default string
	Value   string
}

// GCPTemplateInput extends TemplateInput with pre-marshaled GCP IAM permission lists.
type GCPTemplateInput struct {
	*stacks.TemplateInput

	// GCPProjectID / GCPRegion render as literals in the normal tfvars and as
	// `${{ inputs.gcp_project_id }}` / `${{ inputs.gcp_region }}` CEL references
	// in the Spacelift blueprint. Empty omits the line.
	GCPProjectID string
	GCPRegion    string

	ProvisionPermissions   string
	MaintenancePermissions string
	DeprovisionPermissions string

	ProvisionPredefinedRole   string
	MaintenancePredefinedRole string
	DeprovisionPredefinedRole string

	BreakGlassRoles []GCPRoleTemplateInput
	CustomRoles     []GCPRoleTemplateInput
	InstallInputs   []GCPInstallInputTemplateInput

	AutoGenerateSecrets []string
	Secrets             []GCPSecretTemplateInput
}

func Render(inputs *stacks.TemplateInput) ([]byte, string, error) {
	inputsT, err := template.New("gcp-stack-inputs").Parse(inputsTmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse gcp inputs template")
	}
	secretsT, err := template.New("gcp-stack-secrets").Parse(secretsTmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse gcp secrets template")
	}

	prov, maint, deprov, provPredefined, maintPredefined, deprovPredefined := extractGCPStandardPermissions(inputs.AppCfg)
	breakGlassRoles := extractGCPRolesFromList(inputs.AppCfg.BreakGlassConfig.Roles)
	customRoles := extractGCPRolesFromList(inputs.AppCfg.PermissionsConfig.CustomRoles)

	var installInputs []GCPInstallInputTemplateInput
	if inputs.AppCfg != nil {
		for _, input := range inputs.AppCfg.InputConfig.AppInputs {
			if input.Source == app.AppInputSourceCustomer {
				installInputs = append(installInputs, GCPInstallInputTemplateInput{
					Name:    input.Name,
					Default: input.Default,
					Value:   input.Default,
				})
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
					Value:       s.Default,
				})
			}
		}
	}

	var gcpProjectID, gcpRegion string
	if inputs.Install.GCPAccount != nil {
		gcpProjectID = inputs.Install.GCPAccount.ProjectID
		gcpRegion = inputs.Install.GCPAccount.Region
	}

	gcpInputs := &GCPTemplateInput{
		TemplateInput:             inputs,
		GCPProjectID:              gcpProjectID,
		GCPRegion:                 gcpRegion,
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

	var inputsBuf bytes.Buffer
	if err = inputsT.Execute(&inputsBuf, gcpInputs); err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp inputs template")
	}
	var secretsBuf bytes.Buffer
	if err = secretsT.Execute(&secretsBuf, gcpInputs); err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp secrets template")
	}

	adminTF, err := renderSpaceliftAdminTF(inputs.Install.ID)
	if err != nil {
		return nil, "", err
	}

	// The blueprint surfaces customer install inputs and secrets as blueprint
	// inputs, so render a variant of the tfvars where those values are
	// `${{ inputs.<id> }}` CEL references instead of literals.
	blueprintTfvarsInput := *gcpInputs
	blueprintTfvarsInput.GCPProjectID = "${{ inputs.gcp_project_id }}"
	blueprintTfvarsInput.GCPRegion = "${{ inputs.gcp_region }}"
	blueprintTfvarsInput.InstallInputs = make([]GCPInstallInputTemplateInput, len(installInputs))
	for i, in := range installInputs {
		blueprintTfvarsInput.InstallInputs[i] = GCPInstallInputTemplateInput{
			Name:    in.Name,
			Default: in.Default,
			Value:   fmt.Sprintf("${{ inputs.%s }}", blueprintInstallInputID(in.Name)),
		}
	}
	blueprintTfvarsInput.Secrets = make([]GCPSecretTemplateInput, len(secrets))
	for i, s := range secrets {
		blueprintTfvarsInput.Secrets[i] = s
		blueprintTfvarsInput.Secrets[i].Value = fmt.Sprintf("${{ inputs.%s }}", blueprintSecretID(s.Name))
	}

	var blueprintInputsBuf, blueprintSecretsBuf bytes.Buffer
	if err = inputsT.Execute(&blueprintInputsBuf, &blueprintTfvarsInput); err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp blueprint inputs template")
	}
	if err = secretsT.Execute(&blueprintSecretsBuf, &blueprintTfvarsInput); err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp blueprint secrets template")
	}

	blueprintYAML, err := renderSpaceliftBlueprint(spaceliftBlueprintData{
		InstallID:     inputs.Install.ID,
		InputsTfvars:  blueprintInputsBuf.String(),
		SecretsTfvars: blueprintSecretsBuf.String(),
		GCPProjectID:  gcpProjectID,
		GCPRegion:     gcpRegion,
		InstallInputs: installInputs,
		Secrets:       secrets,
	})
	if err != nil {
		return nil, "", err
	}

	// Wrap tfvars in a JSON envelope so it can be stored in the jsonb column.
	// The raw tfvars text is HCL, not valid JSON.
	envelope := map[string]string{
		"inputs_tfvars":            inputsBuf.String(),
		"secrets_tfvars":           secretsBuf.String(),
		"spacelift_admin_tf":       adminTF,
		"spacelift_blueprint_yaml": blueprintYAML,
	}
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

// GCPOpRoleRaw is a standard operation role's GCP IAM inputs in raw Go form
// (not template strings). Used by the SDK-provisioner config builder.
type GCPOpRoleRaw struct {
	Permissions    []string
	PredefinedRole string
}

// GCPRoleRaw is a named break-glass/custom GCP role in raw Go form.
type GCPRoleRaw struct {
	Name           string
	Permissions    []string
	PredefinedRole string
}

// ExtractGCPStandardRolesRaw returns the provision/maintenance/deprovision GCP
// operation-role inputs (permissions + predefined role) in raw Go form.
func ExtractGCPStandardRolesRaw(appCfg *app.AppConfig) (provision, maintenance, deprovision GCPOpRoleRaw) {
	if appCfg == nil {
		return
	}
	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "gcp" {
			continue
		}
		perms, predefined := extractRolePermissions(role)
		if len(perms) == 0 && predefined == "" {
			continue
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = GCPOpRoleRaw{Permissions: perms, PredefinedRole: predefined}
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = GCPOpRoleRaw{Permissions: perms, PredefinedRole: predefined}
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = GCPOpRoleRaw{Permissions: perms, PredefinedRole: predefined}
		}
	}
	return
}

// ExtractGCPRolesRaw converts a slice of role configs into raw GCP role inputs,
// filtering to GCP roles that have permissions or a predefined role.
func ExtractGCPRolesRaw(roles []app.AppAWSIAMRoleConfig) []GCPRoleRaw {
	var out []GCPRoleRaw
	for _, role := range roles {
		if role.CloudPlatform != "gcp" {
			continue
		}
		perms, predefined := extractRolePermissions(role)
		if len(perms) == 0 && predefined == "" {
			continue
		}
		out = append(out, GCPRoleRaw{Name: role.Name, Permissions: perms, PredefinedRole: predefined})
	}
	return out
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
