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

// GCPPolicyTemplateInput holds one policy rendered as its own custom role,
// mirroring how AWS attaches each policy separately instead of merging.
type GCPPolicyTemplateInput struct {
	Name        string
	Permissions string
}

// GCPRoleTemplateInput holds the per-role data rendered into the template.
type GCPRoleTemplateInput struct {
	Name           string
	Policies       []GCPPolicyTemplateInput
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

	ProvisionPolicies   []GCPPolicyTemplateInput
	MaintenancePolicies []GCPPolicyTemplateInput
	DeprovisionPolicies []GCPPolicyTemplateInput

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
	providerT, err := template.New("gcp-stack-provider-inputs").Parse(providerInputsTmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse gcp provider inputs template")
	}

	prov, maint, deprov, provPredefined, maintPredefined, deprovPredefined := extractGCPStandardPolicies(inputs.AppCfg)
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
		ProvisionPolicies:         prov,
		MaintenancePolicies:       maint,
		DeprovisionPolicies:       deprov,
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
	var providerBuf bytes.Buffer
	if err = providerT.Execute(&providerBuf, gcpInputs); err != nil {
		return nil, "", errors.Wrap(err, "unable to execute gcp provider inputs template")
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
		"provider_tfvars":          providerBuf.String(),
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

// extractGCPStandardPolicies reads GCP IAM policies for the standard roles (provision, maintenance, deprovision).
func extractGCPStandardPolicies(appCfg *app.AppConfig) (provision, maintenance, deprovision []GCPPolicyTemplateInput, provPredefined, maintPredefined, deprovPredefined string) {
	if appCfg == nil {
		return
	}

	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "gcp" {
			continue
		}

		policies, predefinedRole := extractRolePolicies(role)

		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = policies
			provPredefined = predefinedRole
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = policies
			maintPredefined = predefinedRole
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = policies
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

		policies, predefinedRole := extractRolePolicies(role)
		if len(policies) == 0 && predefinedRole == "" {
			continue
		}

		result = append(result, GCPRoleTemplateInput{
			Name:           role.Name,
			Policies:       policies,
			PredefinedRole: predefinedRole,
		})
	}

	return result
}

// extractRolePolicies keeps each policy separate so the stack creates one
// custom role per policy, matching the AWS one-policy-one-attachment shape.
func extractRolePolicies(role app.AppAWSIAMRoleConfig) ([]GCPPolicyTemplateInput, string) {
	var policies []GCPPolicyTemplateInput
	var predefinedRole string
	for i, policy := range role.Policies {
		if policy.GCPPredefinedRole != "" {
			predefinedRole = policy.GCPPredefinedRole
		}

		if len(policy.GCPPermissions) == 0 {
			continue
		}

		b, err := json.Marshal(policy.GCPPermissions)
		if err != nil {
			continue
		}

		name := policy.Name
		if name == "" {
			name = fmt.Sprintf("policy-%d", i)
		}

		policies = append(policies, GCPPolicyTemplateInput{
			Name:        name,
			Permissions: string(b),
		})
	}
	return policies, predefinedRole
}
