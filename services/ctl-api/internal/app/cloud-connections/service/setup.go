package service

import (
	"encoding/json"
	"fmt"
	"net/url"
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
	if connection.Platform == app.CloudPlatformAzure {
		return s.acrSetup(connection, repositories)
	}
	if connection.Platform == app.CloudPlatformGCP {
		return s.garSetup(connection, repositories)
	}
	response := SetupResponse{Subject: subject(connection), Audience: "sts.amazonaws.com", Preset: connection.Preset}
	if s.issuer == nil {
		return response
	}
	issuerURL := s.issuer.Issuer()
	issuer, _ := url.Parse(issuerURL)
	conditionPrefix := strings.TrimPrefix(issuer.Host+issuer.Path, "/")
	trustPolicy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []any{map[string]any{
			"Effect":    "Allow",
			"Principal": map[string]string{"Federated": fmt.Sprintf("arn:aws:iam::%s:oidc-provider/%s", connection.TargetID, conditionPrefix)},
			"Action":    "sts:AssumeRoleWithWebIdentity",
			"Condition": map[string]any{"StringEquals": map[string]string{
				conditionPrefix + ":aud": "sts.amazonaws.com",
				conditionPrefix + ":sub": subject(connection),
			}},
		}},
	}
	policyJSON, _ := json.MarshalIndent(trustPolicy, "", "  ")
	response.IssuerURL = issuerURL
	response.TrustPolicy = trustPolicy
	response.Terraform = fmt.Sprintf(`resource "aws_iam_openid_connect_provider" "nuon" {
  url             = %q
  client_id_list  = ["sts.amazonaws.com"]
}

resource "aws_iam_role" "nuon_cloud_connection" {
  name               = "<role-name>"
  assume_role_policy = jsonencode(%s)
}
`, issuerURL, policyJSON)
	response.CLI = fmt.Sprintf("aws iam create-open-id-connect-provider --url %s --client-id-list sts.amazonaws.com\naws iam create-role --role-name <role-name> --assume-role-policy-document '%s'", issuerURL, policyJSON)
	response.CloudFormation = fmt.Sprintf("NuonOIDCProvider:\n  Type: AWS::IAM::OIDCProvider\n  Properties:\n    Url: %s\n    ClientIdList:\n      - sts.amazonaws.com\nNuonConnectionRole:\n  Type: AWS::IAM::Role\n  Properties:\n    RoleName: <role-name>\n    AssumeRolePolicyDocument:\n      %s\n", issuerURL, strings.ReplaceAll(string(policyJSON), "\n", "\n      "))
	switch connection.Preset {
	case app.CloudConnectionPresetStacks:
		response.PermissionsPolicy = awsPermissionsPolicy()
	case app.CloudConnectionPresetECR:
		if len(repositories) > 0 {
			response.PermissionsPolicy = awsECRPermissionsPolicy(repositories)
			response.PermissionsCLI = awsECRPermissionsCLI(connection, response.PermissionsPolicy)
		}
	}
	if response.PermissionsPolicy != nil {
		permissionsJSON, _ := json.MarshalIndent(response.PermissionsPolicy, "", "  ")
		response.Terraform += fmt.Sprintf(`
resource "aws_iam_role_policy" "nuon_cloud_connection" {
  name   = "nuon-cloud-connection"
  role   = aws_iam_role.nuon_cloud_connection.id
  policy = jsonencode(%s)
}`, permissionsJSON)
		if connection.Preset == app.CloudConnectionPresetECR {
			response.CLI += "\n" + response.PermissionsCLI
		} else {
			response.CLI += fmt.Sprintf("\naws iam put-role-policy --role-name <role-name> --policy-name nuon-cloud-connection --policy-document '%s'", permissionsJSON)
		}
		response.CloudFormation += fmt.Sprintf("    Policies:\n      - PolicyName: nuon-cloud-connection\n        PolicyDocument:\n          %s\n", strings.ReplaceAll(string(permissionsJSON), "\n", "\n          "))
	}
	return response
}

func (s *service) garSetup(connection *app.CloudConnection, repositories []string) SetupResponse {
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
	response.CLI = fmt.Sprintf(`gcloud services enable iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com artifactregistry.googleapis.com --project=%[1]s
gcloud iam service-accounts create %[2]s --project=%[1]s
gcloud iam workload-identity-pools create %[3]s --project=%[1]s --location=global --display-name="Nuon image repository access"
cat > nuon-oidc-jwks.json <<'NUON_JWKS'
%[4]s
NUON_JWKS
gcloud iam workload-identity-pools providers create-oidc nuon --project=%[1]s --location=global --workload-identity-pool=%[3]s --issuer-uri=%[5]s --allowed-audiences=%[6]s --attribute-mapping="google.subject=assertion.sub" --attribute-condition="%[7]s" --jwk-json-path=nuon-oidc-jwks.json
gcloud iam service-accounts add-iam-policy-binding %[8]s --project=%[1]s --role=roles/iam.workloadIdentityUser --member=%[9]s
`, connection.TargetID, saName, poolID, jwks, response.IssuerURL, response.Audience, condition, connection.Principal, member)
	if repositories == nil {
		repositories = []string{}
	}
	repositoriesJSON, _ := json.Marshal(repositories)
	response.Terraform = fmt.Sprintf(`resource "google_project_service" "nuon" {
  for_each = toset(["iam.googleapis.com", "iamcredentials.googleapis.com", "sts.googleapis.com", "artifactregistry.googleapis.com"])
  project = %[1]q
  service = each.value
  disable_on_destroy = false
}
resource "google_service_account" "nuon" {
  project = %[1]q
  account_id = %[2]q
  depends_on = [google_project_service.nuon]
}
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
  service_account_id = google_service_account.nuon.name
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
  member = "serviceAccount:${google_service_account.nuon.email}"
}
`, connection.TargetID, saName, poolID, condition, response.IssuerURL, response.Audience, jwks, member, repositoriesJSON)
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
	roleName := connection.Principal[strings.LastIndex(connection.Principal, "/")+1:]
	return fmt.Sprintf(`set -eu
role_name='%[1]s'
expected_account='%[2]s'
actual_account=$(aws sts get-caller-identity --query Account --output text)
if [ "$actual_account" != "$expected_account" ]; then
  printf 'Select credentials for the configured AWS account before updating permissions.\n' >&2
  exit 1
fi
policy_names=$(aws iam list-role-policies --role-name "$role_name" --output json)
if printf '%%s' "$policy_names" | jq -e '.PolicyNames | index("nuon-image-repositories") != null' >/dev/null; then
  existing=$(aws iam get-role-policy --role-name "$role_name" --policy-name nuon-image-repositories --query PolicyDocument --output json)
else
  existing='{"Version":"2012-10-17","Statement":[]}'
fi
grant=$(cat <<'NUON_REPOSITORY_GRANT'
%[3]s
NUON_REPOSITORY_GRANT
)
merged=$(printf '%%s' "$existing" | jq --argjson grant "$grant" '.Statement = ((if (.Statement | type) == "array" then .Statement else [.Statement] end) + ($grant.Statement | map(del(.Sid))) | unique)')
aws iam put-role-policy --role-name "$role_name" --policy-name nuon-image-repositories --policy-document "$merged"
`, strings.ReplaceAll(roleName, "'", "'\"'\"'"), connection.TargetID, grantJSON)
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
