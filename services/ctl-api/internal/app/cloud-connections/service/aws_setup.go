package service

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

func (s *service) awsSetup(connection *app.CloudConnection, existingPrincipal bool, repositories []string) SetupResponse {
	response := SetupResponse{Subject: cloudconnections.Subject(connection), Audience: "sts.amazonaws.com", Preset: connection.Preset}
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
	response.CLI = awsAccountPreflight(connection, roleName) + fmt.Sprintf(`provider_arn=%[1]s
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
trust=$(cat <<'NUON_CONNECTION_TRUST'
%[4]s
NUON_CONNECTION_TRUST
)
role_error=$(mktemp)
trap 'rm -f "$role_error"' EXIT HUP INT TERM
if role=$(aws iam get-role --role-name "$role_name" --output json 2>"$role_error"); then
  existing=$(printf '%%s' "$role" | jq -e --arg arn %[6]s 'if .Role.Arn == $arn then .Role.AssumeRolePolicyDocument else error("The role ARN does not match the configured principal") end')
  merged=$(printf '%%s' "$existing" | jq -e --argjson trust "$trust" '
    if type != "object" or ((.Statement | type) != "array" and (.Statement | type) != "object") then error("Invalid existing trust policy") else
    .Statement = ((if (.Statement | type) == "array" then .Statement else [.Statement] end) | map(select(.Sid != $trust.Statement[0].Sid))) + $trust.Statement end')
  aws iam update-assume-role-policy --role-name "$role_name" --policy-document "$merged"
elif grep -q NoSuchEntity "$role_error"; then
  aws iam create-role --role-name "$role_name" --path %[5]s --assume-role-policy-document "$trust"
else
  cat "$role_error" >&2
  exit 1
fi
`, shellQuote(providerARN), shellQuote(response.IssuerURL), shellQuote(conditionPrefix), policyJSON, shellQuote(rolePath), shellQuote(connection.Principal))
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
		response.PermissionsCLI = awsPermissionsCLI(connection, roleName, response.PermissionsPolicy)
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

func awsAccountPreflight(connection *app.CloudConnection, roleName string) string {
	return fmt.Sprintf(`set -eu
command -v jq >/dev/null
role_name=%[1]s
expected_account=%[2]s
actual_account=$(aws sts get-caller-identity --query Account --output text)
if [ "$actual_account" != "$expected_account" ]; then
  printf 'Select credentials for the configured AWS account before updating IAM.\n' >&2
  exit 1
fi
`, shellQuote(roleName), shellQuote(connection.TargetID))
}

func awsPermissionsCLI(connection *app.CloudConnection, roleName string, policy map[string]any) string {
	grantJSON, _ := json.MarshalIndent(policy, "", "  ")
	return awsAccountPreflight(connection, roleName) + fmt.Sprintf(`policy_name=%[1]s
role=$(aws iam get-role --role-name "$role_name" --output json)
printf '%%s' "$role" | jq -e --arg arn %[2]s '.Role.Arn == $arn' >/dev/null
policy_names=$(aws iam list-role-policies --role-name "$role_name" --output json)
has_policy=$(printf '%%s' "$policy_names" | jq -er --arg name "$policy_name" 'if (.PolicyNames | type) == "array" then (.PolicyNames | index($name) != null) | tostring else error("Invalid policy list") end')
if [ "$has_policy" = true ]; then
  existing=$(aws iam get-role-policy --role-name "$role_name" --policy-name "$policy_name" --query PolicyDocument --output json)
else
  existing='{"Version":"2012-10-17","Statement":[]}'
fi
grant=$(cat <<'NUON_REPOSITORY_GRANT'
%[3]s
NUON_REPOSITORY_GRANT
)
merged=$(printf '%%s' "$existing" | jq -e --argjson grant "$grant" '
  if type != "object" or ((.Statement | type) != "array" and (.Statement | type) != "object") then error("Invalid existing permissions policy") else
  .Statement = ((if (.Statement | type) == "array" then .Statement else [.Statement] end) + ($grant.Statement | map(del(.Sid))) | unique) end')
aws iam put-role-policy --role-name "$role_name" --policy-name "$policy_name" --policy-document "$merged"
`, shellQuote("nuon-"+connection.ID), shellQuote(connection.Principal), grantJSON)
}

func awsPermissionsPolicy() map[string]any {
	return map[string]any{"Version": "2012-10-17", "Statement": []any{map[string]any{
		"Sid":      "ManageNuonInstallStack",
		"Effect":   "Allow",
		"Action":   awsStackActions,
		"Resource": "*",
	}}}
}

func awsECRPermissionsPolicy(repositoryARNs []string) map[string]any {
	return map[string]any{"Version": "2012-10-17", "Statement": []any{
		map[string]any{
			"Sid": "AuthorizeECR", "Effect": "Allow",
			"Action": []string{"ecr:GetAuthorizationToken"}, "Resource": "*",
		},
		map[string]any{
			"Sid": "ReadECRRepositories", "Effect": "Allow",
			"Action": []string{
				"ecr:DescribeRepositories", "ecr:ListImages", "ecr:BatchGetImage",
				"ecr:GetDownloadUrlForLayer", "ecr:BatchCheckLayerAvailability",
			},
			"Resource": repositoryARNs,
		},
	}}
}

var awsStackActions = []string{
	"autoscaling:CreateAutoScalingGroup", "autoscaling:CreateOrUpdateTags", "autoscaling:DeleteAutoScalingGroup", "autoscaling:DeleteTags", "autoscaling:DescribeAutoScalingGroups", "autoscaling:DescribeAutoScalingInstances", "autoscaling:DescribeLifecycleHooks", "autoscaling:DescribeNotificationConfigurations", "autoscaling:DescribeScalingActivities", "autoscaling:DescribeTags", "autoscaling:ResumeProcesses", "autoscaling:SetDesiredCapacity", "autoscaling:SuspendProcesses", "autoscaling:TerminateInstanceInAutoScalingGroup", "autoscaling:UpdateAutoScalingGroup",
	"cloudformation:CreateStack", "cloudformation:DeleteStack", "cloudformation:DescribeStackEvents", "cloudformation:DescribeStackResource", "cloudformation:DescribeStacks", "cloudformation:GetTemplate", "cloudformation:GetTemplateSummary", "cloudformation:ListStackResources", "cloudformation:ListStacks", "cloudformation:UpdateStack", "cloudformation:ValidateTemplate",
	"ec2:AllocateAddress", "ec2:AssociateRouteTable", "ec2:AttachInternetGateway", "ec2:AuthorizeSecurityGroupEgress", "ec2:AuthorizeSecurityGroupIngress", "ec2:CreateInternetGateway", "ec2:CreateLaunchTemplate", "ec2:CreateLaunchTemplateVersion", "ec2:CreateNatGateway", "ec2:CreateRoute", "ec2:CreateRouteTable", "ec2:CreateSecurityGroup", "ec2:CreateSubnet", "ec2:CreateTags", "ec2:CreateVpc", "ec2:DeleteInternetGateway", "ec2:DeleteLaunchTemplate", "ec2:DeleteNatGateway", "ec2:DeleteRoute", "ec2:DeleteRouteTable", "ec2:DeleteSecurityGroup", "ec2:DeleteSubnet", "ec2:DeleteTags", "ec2:DeleteVpc", "ec2:Describe*", "ec2:DetachInternetGateway", "ec2:DisassociateRouteTable", "ec2:ModifySubnetAttribute", "ec2:ModifyVpcAttribute", "ec2:ReleaseAddress", "ec2:ReplaceRoute", "ec2:ReplaceRouteTableAssociation", "ec2:RevokeSecurityGroupEgress", "ec2:RevokeSecurityGroupIngress", "ec2:RunInstances", "ec2:UpdateSecurityGroupRuleDescriptionsEgress", "ec2:UpdateSecurityGroupRuleDescriptionsIngress",
	"iam:AddRoleToInstanceProfile", "iam:AttachRolePolicy", "iam:CreateInstanceProfile", "iam:CreatePolicy", "iam:CreatePolicyVersion", "iam:CreateRole", "iam:DeleteInstanceProfile", "iam:DeletePolicy", "iam:DeletePolicyVersion", "iam:DeleteRole", "iam:DeleteRolePermissionsBoundary", "iam:DeleteRolePolicy", "iam:DetachRolePolicy", "iam:GetAccountSummary", "iam:GetInstanceProfile", "iam:GetPolicy", "iam:GetPolicyVersion", "iam:GetRole", "iam:GetRolePolicy", "iam:ListAttachedRolePolicies", "iam:ListEntitiesForPolicy", "iam:ListInstanceProfilesForRole", "iam:ListPoliciesGrantingServiceAccess", "iam:ListPolicyVersions", "iam:ListRolePolicies", "iam:PassRole", "iam:PutRolePermissionsBoundary", "iam:PutRolePolicy", "iam:RemoveRoleFromInstanceProfile", "iam:TagInstanceProfile", "iam:TagPolicy", "iam:TagRole", "iam:UntagInstanceProfile", "iam:UntagPolicy", "iam:UntagRole", "iam:UpdateAssumeRolePolicy", "iam:UpdateRole", "iam:UpdateRoleDescription",
	"lambda:AddPermission", "lambda:CreateFunction", "lambda:DeleteFunction", "lambda:GetAccountSettings", "lambda:GetFunction", "lambda:GetFunctionCodeSigningConfig", "lambda:GetFunctionConfiguration", "lambda:GetFunctionRecursionConfig", "lambda:GetFunctionScalingConfig", "lambda:GetRuntimeManagementConfig", "lambda:InvokeFunction", "lambda:RemovePermission", "lambda:TagResource", "lambda:UntagResource", "lambda:UpdateFunctionCode", "lambda:UpdateFunctionConfiguration",
	"logs:CreateLogGroup", "logs:CreateLogStream", "logs:DeleteLogGroup", "logs:DeleteLogStream", "logs:DeleteRetentionPolicy", "logs:DescribeIndexPolicies", "logs:DescribeLogGroups", "logs:DescribeLogStreams", "logs:DescribeResourcePolicies", "logs:GetDataProtectionPolicy", "logs:ListTagsForResource", "logs:PutRetentionPolicy", "logs:TagResource", "logs:UntagResource",
	"route53resolver:AssociateFirewallRuleGroup", "route53resolver:CreateFirewallDomainList", "route53resolver:CreateFirewallRuleGroup", "route53resolver:CreateFirewallRule", "route53resolver:DeleteFirewallDomainList", "route53resolver:DeleteFirewallRuleGroup", "route53resolver:DeleteFirewallRule", "route53resolver:DisassociateFirewallRuleGroup", "route53resolver:GetFirewallConfig", "route53resolver:GetFirewallDomainList", "route53resolver:GetFirewallRuleGroup", "route53resolver:GetFirewallRuleGroupAssociation", "route53resolver:ListFirewallDomainLists", "route53resolver:ListFirewallDomains", "route53resolver:ListFirewallRuleGroupAssociations", "route53resolver:ListFirewallRuleGroups", "route53resolver:ListFirewallRules", "route53resolver:ListTagsForResource", "route53resolver:TagResource", "route53resolver:UntagResource", "route53resolver:UpdateFirewallDomains", "route53resolver:UpdateFirewallRule", "route53resolver:UpdateFirewallRuleGroupAssociation",
	"s3:GetObject",
	"secretsmanager:CreateSecret", "secretsmanager:DeleteSecret", "secretsmanager:DescribeSecret", "secretsmanager:GetRandomPassword", "secretsmanager:GetResourcePolicy", "secretsmanager:PutResourcePolicy", "secretsmanager:TagResource", "secretsmanager:UntagResource", "secretsmanager:UpdateSecret",
	"servicequotas:GetServiceQuota",
	"ssm:GetParameters",
	"iam:CreateServiceLinkedRole",
}
