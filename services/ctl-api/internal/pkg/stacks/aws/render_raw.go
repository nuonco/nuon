package aws

import (
	"encoding/json"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// AWSRoleRaw is the un-rendered per-role payload the stack SDK consumes: an
// operation/break-glass/custom role's merged inline policy and managed policy
// ARNs, without the CloudFormation-specific wrapping.
type AWSRoleRaw struct {
	Name                 string
	InlinePolicyDocument string
	ManagedPolicyARNs    []string
}

// ExtractAWSStandardPermissionsRaw returns the managed policy ARNs for the
// standard provision/maintenance/deprovision operation roles.
func ExtractAWSStandardPermissionsRaw(appCfg *app.AppConfig) (provision, maintenance, deprovision []string) {
	if appCfg == nil {
		return nil, nil, nil
	}
	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}
		mpas := managedPolicyArnsForRole(role)
		if len(mpas) == 0 {
			continue
		}
		switch role.Type {
		case app.AWSIAMRoleTypeRunnerProvision:
			provision = mpas
		case app.AWSIAMRoleTypeRunnerMaintenance:
			maintenance = mpas
		case app.AWSIAMRoleTypeRunnerDeprovision:
			deprovision = mpas
		}
	}
	return provision, maintenance, deprovision
}

// ExtractAWSStandardInlinePoliciesRaw returns the merged inline policy document
// for each standard operation role.
func ExtractAWSStandardInlinePoliciesRaw(appCfg *app.AppConfig) (provision, maintenance, deprovision string, err error) {
	if appCfg == nil {
		return "", "", "", nil
	}
	for _, role := range appCfg.PermissionsConfig.Roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}
		doc, derr := mergedInlinePolicyDocumentRaw(role)
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

// ExtractAWSRolesFromListRaw returns the raw payload for a list of break-glass
// or custom AWS roles, skipping non-AWS and empty roles.
func ExtractAWSRolesFromListRaw(roles []app.AppAWSIAMRoleConfig) ([]AWSRoleRaw, error) {
	var result []AWSRoleRaw
	for _, role := range roles {
		if role.CloudPlatform != "" && role.CloudPlatform != "aws" {
			continue
		}
		mpas := managedPolicyArnsForRole(role)
		doc, err := mergedInlinePolicyDocumentRaw(role)
		if err != nil {
			return nil, fmt.Errorf("role %q: %w", role.Name, err)
		}
		if len(mpas) == 0 && doc == "" {
			continue
		}
		result = append(result, AWSRoleRaw{
			Name:                 role.Name,
			InlinePolicyDocument: doc,
			ManagedPolicyARNs:    mpas,
		})
	}
	return result, nil
}

// mergedInlinePolicyDocumentRaw merges a role's inline policy statements into a
// single IAM policy document, excluding managed-policy references.
func mergedInlinePolicyDocumentRaw(role app.AppAWSIAMRoleConfig) (string, error) {
	var statements []json.RawMessage
	for _, policy := range role.Policies {
		if len(policy.Contents) == 0 {
			continue
		}
		if policy.ManagedPolicyName != "" {
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
	return string(b), nil
}
