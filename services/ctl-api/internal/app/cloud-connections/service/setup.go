package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type SetupResponse struct {
	IssuerURL         string                    `json:"issuer_url"`
	Subject           string                    `json:"subject"`
	Audience          string                    `json:"audience"`
	TrustPolicy       map[string]any            `json:"trust_policy"`
	PermissionsPolicy map[string]any            `json:"permissions_policy,omitempty"`
	Terraform         string                    `json:"terraform"`
	CLI               string                    `json:"cli"`
	CloudFormation    string                    `json:"cloudformation"`
	Preset            app.CloudConnectionPreset `json:"preset"`
}

func subject(connection *app.CloudConnection) string {
	return fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID)
}

func (s *service) setup(connection *app.CloudConnection) SetupResponse {
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
	if connection.Preset == app.CloudConnectionPresetStacks {
		response.PermissionsPolicy = awsPermissionsPolicy()
		permissionsJSON, _ := json.MarshalIndent(response.PermissionsPolicy, "", "  ")
		response.Terraform += fmt.Sprintf(`
resource "aws_iam_role_policy" "nuon_cloud_connection" {
  name   = "nuon-cloud-connection"
  role   = aws_iam_role.nuon_cloud_connection.id
  policy = jsonencode(%s)
}`, permissionsJSON)
		response.CLI += fmt.Sprintf("\naws iam put-role-policy --role-name <role-name> --policy-name nuon-cloud-connection --policy-document '%s'", permissionsJSON)
		response.CloudFormation += fmt.Sprintf("    Policies:\n      - PolicyName: nuon-cloud-connection\n        PolicyDocument:\n          %s\n", strings.ReplaceAll(string(permissionsJSON), "\n", "\n          "))
	}
	return response
}

func awsPermissionsPolicy() map[string]any {
	return map[string]any{"Version": "2012-10-17", "Statement": []any{map[string]any{
		"Sid":      "ManageNuonInstallStack",
		"Effect":   "Allow",
		"Action":   awsStackActions,
		"Resource": "*",
	}}}
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
	"secretsmanager:CreateSecret", "secretsmanager:DeleteSecret", "secretsmanager:DescribeSecret", "secretsmanager:GetResourcePolicy", "secretsmanager:PutResourcePolicy", "secretsmanager:TagResource", "secretsmanager:UntagResource", "secretsmanager:UpdateSecret",
	"servicequotas:GetServiceQuota",
	"ssm:GetParameters",
	"iam:CreateServiceLinkedRole",
}
