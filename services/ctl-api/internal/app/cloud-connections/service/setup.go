package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	cloudconnections "github.com/nuonco/nuon/services/ctl-api/internal/app/cloud-connections"
)

type SetupResponse struct {
	IssuerURL         string                    `json:"issuer_url"`
	Subject           string                    `json:"subject"`
	Audience          string                    `json:"audience"`
	TrustPolicy       map[string]any            `json:"trust_policy"`
	PermissionsPolicy map[string]any            `json:"permissions_policy,omitempty"`
	PermissionsCLI    string                    `json:"permissions_cli,omitempty"`
	Terraform         string                    `json:"terraform"`
	CLI               string                    `json:"cli"`
	CloudFormation    string                    `json:"cloudformation"`
	Preset            app.CloudConnectionPreset `json:"preset"`
}

func subject(connection *app.CloudConnection) string {
	return fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID)
}

func (s *service) setup(connection *app.CloudConnection, repositories ...string) SetupResponse {
	return s.setupWithPrincipal(connection, false, repositories...)
}

func (s *service) setupWithPrincipal(connection *app.CloudConnection, existingPrincipal bool, repositories ...string) SetupResponse {
	if connection.Platform == app.CloudPlatformAzure {
		return s.acrSetup(connection, existingPrincipal, repositories)
	}
	if connection.Platform == app.CloudPlatformGCP {
		return s.garSetup(connection, existingPrincipal, repositories)
	}
	return s.awsSetup(connection, existingPrincipal, repositories)
}

func (s *service) garSetup(connection *app.CloudConnection, existingPrincipal bool, repositories []string) SetupResponse {
	response := SetupResponse{Subject: subject(connection), Audience: "//iam.googleapis.com/" + cloudconnections.GARProvider(connection), Preset: connection.Preset}
	if s.issuer == nil {
		return response
	}
	response.IssuerURL = s.issuer.Issuer()
	poolID := "nuon-" + connection.ID
	pool := fmt.Sprintf("projects/%s/locations/global/workloadIdentityPools/%s", connection.GCPProjectNumber, poolID)
	member := "principal://iam.googleapis.com/" + pool + "/subject/" + response.Subject
	condition := "assertion.sub == '" + response.Subject + "'"
	response.TrustPolicy = map[string]any{
		"issuer_uri": response.IssuerURL, "allowed_audiences": []string{response.Audience},
		"attribute_mapping":   map[string]string{"google.subject": "assertion.sub"},
		"attribute_condition": condition,
		"service_account":     connection.Principal, "role": "roles/iam.workloadIdentityUser", "member": member,
	}
	bindings := make([]map[string]string, 0, len(repositories))
	for _, repository := range repositories {
		bindings = append(bindings, map[string]string{"resource": repository, "role": "roles/artifactregistry.reader", "member": "serviceAccount:" + connection.Principal})
	}
	if len(bindings) > 0 {
		response.PermissionsPolicy = map[string]any{"bindings": bindings}
	}
	jwks, _ := json.MarshalIndent(s.issuer.PublicJWKS(), "", "  ")
	saName := strings.Split(connection.Principal, "@")[0]
	response.CLI = fmt.Sprintf(`set -eu
command -v jq >/dev/null
project=%[1]s
service_account=%[8]s
pool_id=%[3]s
pool_name=%[10]s
project_number=$(gcloud projects describe "$project" --format='value(projectNumber)')
if [ "$project_number" != %[11]s ]; then
  printf 'The GCP project number must match the configured project.\n' >&2
  exit 1
fi
gcloud services enable iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com artifactregistry.googleapis.com --project="$project"
accounts=$(gcloud iam service-accounts list --project="$project" --format=json)
account_count=$(printf '%%s' "$accounts" | jq -er --arg email "$service_account" 'if type == "array" then [.[] | select(.email == $email)] | length else error("Invalid service account list") end')
if [ "$account_count" = 0 ]; then
  gcloud iam service-accounts create %[2]s --project="$project"
fi
pools=$(gcloud iam workload-identity-pools list --project="$project" --location=global --format=json)
pool_count=$(printf '%%s' "$pools" | jq -er --arg name "$pool_name" 'if type == "array" then [.[] | select(.name == $name)] | length else error("Invalid pool list") end')
if [ "$pool_count" = 0 ]; then
  gcloud iam workload-identity-pools create "$pool_id" --project="$project" --location=global --display-name="Nuon image repository access"
fi
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT HUP INT TERM
cat > "$workdir/jwks.json" <<'NUON_JWKS'
%[4]s
NUON_JWKS
providers=$(gcloud iam workload-identity-pools providers list --project="$project" --location=global --workload-identity-pool="$pool_id" --format=json)
provider=$(printf '%%s' "$providers" | jq -ce --arg name "$pool_name/providers/nuon" 'if type == "array" then [.[] | select(.name == $name)] else error("Invalid provider list") end')
if [ "$(printf '%%s' "$provider" | jq 'length')" = 0 ]; then
  provider_operation=create-oidc
else
  printf '%%s' "$provider" | jq -e --arg issuer %[5]s --arg audience %[6]s --arg condition %[7]s 'length == 1 and .[0].oidc.issuerUri == $issuer and .[0].oidc.allowedAudiences == [$audience] and .[0].attributeCondition == $condition and .[0].attributeMapping["google.subject"] == "assertion.sub"' >/dev/null
  provider_operation=update-oidc
fi
gcloud iam workload-identity-pools providers "$provider_operation" nuon --project="$project" --location=global --workload-identity-pool="$pool_id" --issuer-uri=%[5]s --allowed-audiences=%[6]s --attribute-mapping="google.subject=assertion.sub" --attribute-condition=%[7]s --jwk-json-path="$workdir/jwks.json"
gcloud iam service-accounts add-iam-policy-binding "$service_account" --project="$project" --role=roles/iam.workloadIdentityUser --member=%[9]s --condition=None
`, acrShellQuote(connection.TargetID), acrShellQuote(saName), acrShellQuote(poolID), jwks, acrShellQuote(response.IssuerURL), acrShellQuote(response.Audience), acrShellQuote(condition), acrShellQuote(connection.Principal), acrShellQuote(member), acrShellQuote(pool), acrShellQuote(connection.GCPProjectNumber))
	if repositories == nil {
		repositories = []string{}
	}
	repositoriesJSON, _ := json.Marshal(repositories)
	serviceAccountType := "resource"
	serviceAccountReference := "google_service_account.nuon"
	if existingPrincipal {
		serviceAccountType = "data"
		serviceAccountReference = "data.google_service_account.nuon"
	}
	serviceAccount := fmt.Sprintf(`%s "google_service_account" "nuon" {
  project = %q
  account_id = %q
  depends_on = [google_project_service.nuon]
}
`, serviceAccountType, connection.TargetID, saName)
	response.Terraform = fmt.Sprintf(`resource "google_project_service" "nuon" {
  for_each = toset(["iam.googleapis.com", "iamcredentials.googleapis.com", "sts.googleapis.com", "artifactregistry.googleapis.com"])
  project = %[1]q
  service = each.value
  disable_on_destroy = false
}
%[10]s
resource "google_iam_workload_identity_pool" "nuon" {
  project = %[1]q
  workload_identity_pool_id = %[3]q
  depends_on = [google_project_service.nuon]
}
resource "google_iam_workload_identity_pool_provider" "nuon" {
  project = %[1]q
  workload_identity_pool_id = google_iam_workload_identity_pool.nuon.workload_identity_pool_id
  workload_identity_pool_provider_id = "nuon"
  attribute_mapping = { "google.subject" = "assertion.sub" }
  attribute_condition = %[4]q
  oidc {
    issuer_uri = %[5]q
    allowed_audiences = [%[6]q]
    jwks_json = jsonencode(%[7]s)
  }
}
resource "google_service_account_iam_member" "nuon" {
  service_account_id = %[11]s.name
  role = "roles/iam.workloadIdentityUser"
  member = %[8]q
  depends_on = [google_iam_workload_identity_pool_provider.nuon]
}
resource "google_artifact_registry_repository_iam_member" "nuon" {
  for_each = toset(%[9]s)
  project = split("/", each.value)[1]
  location = split("/", each.value)[3]
  repository = split("/", each.value)[5]
  role = "roles/artifactregistry.reader"
  member = "serviceAccount:${%[11]s.email}"
}
`, connection.TargetID, saName, poolID, condition, response.IssuerURL, response.Audience, jwks, member, repositoriesJSON, serviceAccount, serviceAccountReference)
	for _, repository := range repositories {
		parts := strings.Split(repository, "/")
		response.PermissionsCLI += fmt.Sprintf("gcloud artifacts repositories add-iam-policy-binding %s --project=%s --location=%s --role=roles/artifactregistry.reader --member=serviceAccount:%s --condition=None\n", parts[5], connection.TargetID, parts[3], connection.Principal)
	}
	if response.PermissionsCLI != "" {
		response.PermissionsCLI = "set -eu\n" + response.PermissionsCLI
		response.CLI += response.PermissionsCLI
	}
	return response
}

func awsECRPermissionsCLI(connection *app.CloudConnection, policy map[string]any) string {
	grantJSON, _ := json.MarshalIndent(policy, "", "  ")
	return awsAccountPreflight(connection) + fmt.Sprintf(`policy_name=%[1]s
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
`, acrShellQuote("nuon-"+connection.ID), acrShellQuote(connection.Principal), grantJSON)
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
