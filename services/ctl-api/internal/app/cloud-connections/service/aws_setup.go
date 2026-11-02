package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *service) awsSetup(connection *app.CloudConnection, existingPrincipal bool, repositories []string) SetupResponse {
	response := SetupResponse{Subject: subject(connection), Audience: "sts.amazonaws.com", Preset: connection.Preset}
	if s.issuer == nil {
		return response
	}
	response.IssuerURL = s.issuer.Issuer()
	issuer, _ := url.Parse(response.IssuerURL)
	conditionPrefix := strings.TrimPrefix(issuer.Host+issuer.Path, "/")
	providerARN := fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", connection.TargetID, conditionPrefix)
	statement := map[string]any{
		"Sid":       fmt.Sprintf("NuonConnection%x", sha256.Sum256([]byte(connection.ID))),
		"Effect":    "Allow",
		"Principal": map[string]string{"Federated": providerARN},
		"Action":    "sts:AssumeRoleWithWebIdentity",
		"Condition": map[string]any{"StringEquals": map[string]string{
			conditionPrefix + ":aud": response.Audience,
			conditionPrefix + ":sub": response.Subject,
		}},
	}
	response.TrustPolicy = map[string]any{"Version": "2012-10-17", "Statement": []any{statement}}
	policyJSON, _ := json.MarshalIndent(response.TrustPolicy, "", "  ")
	statementJSON, _ := json.MarshalIndent(statement, "", "  ")
	roleResource := strings.TrimPrefix(connection.Principal, "arn:aws:iam::"+connection.TargetID+":role/")
	roleName := roleResource[strings.LastIndex(roleResource, "/")+1:]
	rolePath := "/" + strings.TrimSuffix(roleResource, roleName)
	response.CLI = awsAccountPreflight(connection) + fmt.Sprintf(`provider_arn=%[1]s
providers=$(aws iam list-open-id-connect-providers --output json)
provider=$(printf '%%s' "$providers" | jq -er --arg arn "$provider_arn" 'if (.OpenIDConnectProviderList | type) == "array" then [.OpenIDConnectProviderList[] | select(.Arn == $arn)] | length else error("Invalid OIDC provider list") end')
if [ "$provider" = 0 ]; then
  aws iam create-open-id-connect-provider --url %[2]s --client-id-list sts.amazonaws.com
else
  provider=$(aws iam get-open-id-connect-provider --open-id-connect-provider-arn "$provider_arn" --output json)
  has_audience=$(printf '%%s' "$provider" | jq -er --arg url %[3]s 'if .Url == $url and (.ClientIDList | type) == "array" then (.ClientIDList | index("sts.amazonaws.com") != null) | tostring else error("Invalid OIDC provider") end')
  if [ "$has_audience" = false ]; then
    aws iam add-client-id-to-open-id-connect-provider --open-id-connect-provider-arn "$provider_arn" --client-id sts.amazonaws.com
  fi
fi
roles=$(aws iam list-roles --output json)
role=$(printf '%%s' "$roles" | jq -er --arg name "$role_name" 'if (.Roles | type) == "array" then [.Roles[] | select(.RoleName == $name)] | length else error("Invalid role list") end')
trust=$(cat <<'NUON_CONNECTION_TRUST'
%[4]s
NUON_CONNECTION_TRUST
)
if [ "$role" = 0 ]; then
  aws iam create-role --role-name "$role_name" --path %[5]s --assume-role-policy-document "$trust"
else
  role=$(aws iam get-role --role-name "$role_name" --output json)
  existing=$(printf '%%s' "$role" | jq -e --arg arn %[6]s 'if .Role.Arn == $arn then .Role.AssumeRolePolicyDocument else error("The role ARN does not match the configured principal") end')
  merged=$(printf '%%s' "$existing" | jq -e --argjson trust "$trust" '
    if type != "object" or ((.Statement | type) != "array" and (.Statement | type) != "object") then error("Invalid existing trust policy") else
    .Statement = ((if (.Statement | type) == "array" then .Statement else [.Statement] end) | map(select(.Sid != $trust.Statement[0].Sid))) + $trust.Statement end')
  aws iam update-assume-role-policy --role-name "$role_name" --policy-document "$merged"
fi
`, acrShellQuote(providerARN), acrShellQuote(response.IssuerURL), acrShellQuote(conditionPrefix), policyJSON, acrShellQuote(rolePath), acrShellQuote(connection.Principal))
	resources := map[string]any{}
	cloudFormation := map[string]any{"AWSTemplateFormatVersion": "2010-09-09", "Resources": resources}
	roleReference := "aws_iam_role.nuon_cloud_connection.id"
	var cfRole any = map[string]string{"Ref": "NuonConnectionRole"}
	if existingPrincipal {
		response.Terraform = fmt.Sprintf(`data "aws_iam_role" "nuon_cloud_connection" {
  name = %[1]q
  lifecycle {
    postcondition {
      condition = self.arn == %[2]q
      error_message = "The existing role ARN must match the configured principal."
    }
  }
}

locals {
  nuon_trust_statement_to_merge = %[3]s
}

output "nuon_trust_statement_to_merge" {
  description = "Merge this exact statement into the existing role's assume-role policy in its original owning Terraform/CloudFormation state, preserving all other statements. Ensure that state owns the issuer OIDC provider with sts.amazonaws.com audience. This configuration does not manage the role or OIDC provider."
  value = jsonencode(local.nuon_trust_statement_to_merge)
}
`, roleName, connection.Principal, statementJSON)
		roleReference = "data.aws_iam_role.nuon_cloud_connection.id"
		cfRole = roleName
		cloudFormation["Description"] = "Additive permissions only. Merge the TrustStatementToMerge output into the existing role in its original owning stack, preserving all other statements. Ensure its issuer OIDC provider allows sts.amazonaws.com. No role or OIDC provider is managed here. If Resources is empty, this is trust-only merge material, not a deployable stack."
		cloudFormation["Outputs"] = map[string]any{"TrustStatementToMerge": map[string]any{
			"Description": "Exact trust statement contribution to merge into the original owning role, not a replacement trust policy.",
			"Value":       string(statementJSON),
		}}
	} else {
		response.Terraform = fmt.Sprintf(`resource "aws_iam_openid_connect_provider" "nuon" {
  url = %[1]q
  client_id_list = ["sts.amazonaws.com"]
}

resource "aws_iam_role" "nuon_cloud_connection" {
  name = %[2]q
  path = %[3]q
  assume_role_policy = jsonencode(%[4]s)
  depends_on = [aws_iam_openid_connect_provider.nuon]
}
`, response.IssuerURL, roleName, rolePath, policyJSON)
		resources["NuonOIDCProvider"] = map[string]any{"Type": "AWS::IAM::OIDCProvider", "Properties": map[string]any{
			"Url": response.IssuerURL, "ClientIdList": []string{response.Audience},
		}}
		resources["NuonConnectionRole"] = map[string]any{"Type": "AWS::IAM::Role", "DependsOn": "NuonOIDCProvider", "Properties": map[string]any{
			"RoleName": roleName, "Path": rolePath, "AssumeRolePolicyDocument": response.TrustPolicy,
		}}
	}
	switch connection.Preset {
	case app.CloudConnectionPresetStacks:
		response.PermissionsPolicy = awsPermissionsPolicy()
	case app.CloudConnectionPresetECR:
		if len(repositories) > 0 {
			response.PermissionsPolicy = awsECRPermissionsPolicy(repositories)
		}
	}
	if response.PermissionsPolicy != nil {
		permissionsJSON, _ := json.MarshalIndent(response.PermissionsPolicy, "", "  ")
		policyName := "nuon-" + connection.ID
		response.PermissionsCLI = awsECRPermissionsCLI(connection, response.PermissionsPolicy)
		response.CLI += response.PermissionsCLI
		response.Terraform += fmt.Sprintf(`
resource "aws_iam_role_policy" "nuon_cloud_connection" {
  name = %[1]q
  role = %[2]s
  policy = jsonencode(%[3]s)
}
`, policyName, roleReference, permissionsJSON)
		resources["NuonConnectionPolicy"] = map[string]any{"Type": "AWS::IAM::Policy", "Properties": map[string]any{
			"PolicyName": policyName, "Roles": []any{cfRole}, "PolicyDocument": response.PermissionsPolicy,
		}}
	}
	cfJSON, _ := json.MarshalIndent(cloudFormation, "", "  ")
	response.CloudFormation = string(cfJSON)
	return response
}

func awsAccountPreflight(connection *app.CloudConnection) string {
	roleName := connection.Principal[strings.LastIndex(connection.Principal, "/")+1:]
	return fmt.Sprintf(`set -eu
command -v jq >/dev/null
role_name=%[1]s
expected_account=%[2]s
actual_account=$(aws sts get-caller-identity --query Account --output text)
if [ "$actual_account" != "$expected_account" ]; then
  printf 'Select credentials for the configured AWS account before updating IAM.\n' >&2
  exit 1
fi
`, acrShellQuote(roleName), acrShellQuote(connection.TargetID))
}
