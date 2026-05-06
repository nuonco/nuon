// Package aws renders the install-stacks/aws Terraform module's tfvars file
// for an AWS install.
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
	Name                 string
	Permissions          string
	InlinePolicyDocument string
	ManagedPolicyArns    string
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

	ProvisionInlinePolicyDocument   string
	MaintenanceInlinePolicyDocument string
	DeprovisionInlinePolicyDocument string

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
// `supportIAMRoleARN` is the Nuon control-plane IAM role ARN that the
// operation roles (provision/maintenance/deprovision/break-glass/custom) must
// trust. Sourced from the ctl-api `runner_default_support_iam_role_arn`
// config — same value the CFN role-builder uses
//
// Custom nested stacks (CloudFormation customer extensions) are intentionally
// not translated. Vendors who extend their CFN stack with custom resources are
// expected to fork install-stacks and make equivalent Terraform changes there.
func Render(inputs *stacks.TemplateInput, supportIAMRoleARN string) ([]byte, string, error) {
	t, err := template.New("aws-stack").Parse(tmpl)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to parse aws template")
	}

	prov, maint, deprov, provMPAs, maintMPAs, deprovMPAs := extractAWSStandardPermissions(inputs.AppCfg)
	provDoc, maintDoc, deprovDoc, err := extractAWSStandardInlinePolicies(inputs.AppCfg)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to extract aws inline policies")
	}
	breakGlassRoles, err := extractAWSRolesFromList(inputs.AppCfg.BreakGlassConfig.Roles)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to extract aws break-glass roles")
	}
	customRoles, err := extractAWSRolesFromList(inputs.AppCfg.PermissionsConfig.CustomRoles)
	if err != nil {
		return nil, "", errors.Wrap(err, "unable to extract aws custom roles")
	}

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

	// Trust principals for the operation roles: the Nuon control-plane support
	// IAM role (when configured), serialized as a JSON list literal for HCL.
	trustPrincipals := []string{}
	if supportIAMRoleARN != "" {
		trustPrincipals = append(trustPrincipals, supportIAMRoleARN)
	}
	trustPrincipalsJSON := "[]"
	if b, err := json.Marshal(trustPrincipals); err == nil {
		trustPrincipalsJSON = string(b)
	}

	awsInputs := &AWSTemplateInput{
		TemplateInput:                   inputs,
		ControlPlaneAccountIDs:          trustPrincipalsJSON,
		ProvisionPermissions:            prov,
		MaintenancePermissions:          maint,
		DeprovisionPermissions:          deprov,
		ProvisionInlinePolicyDocument:   provDoc,
		MaintenanceInlinePolicyDocument: maintDoc,
		DeprovisionInlinePolicyDocument: deprovDoc,
		ProvisionManagedPolicyArns:      provMPAs,
		MaintenanceManagedPolicyArns:    maintMPAs,
		DeprovisionManagedPolicyArns:    deprovMPAs,
		BreakGlassRoles:                 breakGlassRoles,
		CustomRoles:                     customRoles,
		InstallInputs:                   installInputs,
		AutoGenerateSecrets:             autoGenerateSecrets,
		Secrets:                         secrets,
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

// extractAWSStandardPermissions reads AWS IAM managed-policy attachments for
// the standard runner roles. Inline policy contents are handled separately by
// extractAWSStandardInlinePolicies.
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

// extractAWSStandardInlinePolicies merges every inline policy document
// (`policy.Contents`) attached to each standard runner role into a single IAM
// policy document per role and returns it as an HCL string literal ready for
// the template (or `""` if no inline policy applies). Mirrors the CFN renderer
// which embeds `Contents` verbatim as `PolicyDocument`.
func extractAWSStandardInlinePolicies(appCfg *app.AppConfig) (provision, maintenance, deprovision string, err error) {
	if appCfg == nil {
		return "", "", "", nil
	}

	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}
		doc, derr := mergedInlinePolicyDocument(role)
		if derr != nil {
			return "", "", "", fmt.Errorf("role %q: %w", role.Name, derr)
		}
		if doc == "" {
			continue
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = doc
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = doc
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = doc
		}
	}

	return provision, maintenance, deprovision, nil
}

// extractAWSRolesFromList converts a slice of role configs into template-ready
// inputs, filtering to AWS roles only. Roles with neither managed-policy
// attachments nor inline policy contents are skipped.
func extractAWSRolesFromList(roles []app.AppAWSIAMRoleConfig) ([]AWSRoleTemplateInput, error) {
	var result []AWSRoleTemplateInput
	for _, role := range roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}
		mpas := managedPolicyArnsForRole(role)
		doc, err := mergedInlinePolicyDocument(role)
		if err != nil {
			return nil, fmt.Errorf("role %q: %w", role.Name, err)
		}
		if len(mpas) == 0 && doc == "" {
			continue
		}
		mpasJSON := "[]"
		if len(mpas) > 0 {
			b, err := json.Marshal(mpas)
			if err != nil {
				return nil, fmt.Errorf("role %q: marshal managed policy arns: %w", role.Name, err)
			}
			mpasJSON = string(b)
		}
		result = append(result, AWSRoleTemplateInput{
			Name:                 role.Name,
			Permissions:          "[]",
			InlinePolicyDocument: doc,
			ManagedPolicyArns:    mpasJSON,
		})
	}
	return result, nil
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

// mergedInlinePolicyDocument merges the Statement arrays from every inline
// policy attached to a role into a single IAM policy document. Returns the
// document JSON-marshaled and HCL-quoted (i.e. ready to interpolate into
// tfvars), or `""` if the role has no inline policy contents.
//
// Each policy.Contents is expected to be a full IAM policy JSON document of
// the form `{"Version": "...", "Statement": [...]}`. Non-Contents policies
// (i.e. ManagedPolicyName-only entries) are skipped.
func mergedInlinePolicyDocument(role app.AppAWSIAMRoleConfig) (string, error) {
	var statements []json.RawMessage
	for _, policy := range role.Policies {
		if len(policy.Contents) == 0 {
			continue
		}
		if policy.ManagedPolicyName != "" {
			// Mutually exclusive with Contents per existing config validation.
			continue
		}
		var doc struct {
			Statement []json.RawMessage `json:"Statement"`
		}
		if err := json.Unmarshal(policy.Contents, &doc); err != nil {
			return "", fmt.Errorf("policy %q: parse inline policy JSON: %w", policy.Name, err)
		}
		statements = append(statements, doc.Statement...)
	}
	if len(statements) == 0 {
		return "", nil
	}
	merged := struct {
		Version   string            `json:"Version"`
		Statement []json.RawMessage `json:"Statement"`
	}{
		Version:   "2012-10-17",
		Statement: statements,
	}
	b, err := json.Marshal(merged)
	if err != nil {
		return "", fmt.Errorf("marshal merged policy document: %w", err)
	}
	q, err := json.Marshal(string(b))
	if err != nil {
		return "", fmt.Errorf("hcl-quote policy document: %w", err)
	}
	return string(q), nil
}
