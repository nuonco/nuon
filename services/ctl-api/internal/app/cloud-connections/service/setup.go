package service

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const azureTokenExchangeAudience = "api://AzureADTokenExchange"

type SetupOptions struct {
	Repositories []string
	Registry     string
}

type SetupResponse struct {
	IssuerURL         string                          `json:"issuer_url"`
	Subject           string                          `json:"subject"`
	Audience          string                          `json:"audience"`
	TrustPolicy       map[string]any                  `json:"trust_policy"`
	PermissionsPolicy map[string]any                  `json:"permissions_policy"`
	Terraform         string                          `json:"terraform"`
	CLI               string                          `json:"cli"`
	PortalJSON        string                          `json:"portal_json,omitempty"`
	CloudFormation    string                          `json:"cloudformation"`
	Capabilities      []app.CloudConnectionCapability `json:"capabilities"`
	Repositories      []string                        `json:"repositories,omitempty"`
	Registry          string                          `json:"registry,omitempty"`
}

func subject(connection *app.CloudConnection) string {
	return fmt.Sprintf("org:%s:connection:%s", connection.OrgID, connection.ID)
}

func (s *service) setup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	switch connection.Platform {
	case app.CloudPlatformAzure:
		return s.azureSetup(connection, options)
	case app.CloudPlatformGCP:
		return s.gcpSetup(connection, options)
	default:
		return s.awsSetup(connection, options)
	}
}

func (s *service) awsSetup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	response := SetupResponse{Subject: subject(connection), Audience: "sts.amazonaws.com", Capabilities: connection.RequestedCapabilities, Repositories: options.Repositories}
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
	permissionsPolicy := awsPermissionsPolicy(connection, options.Repositories)
	policyJSON, _ := json.Marshal(trustPolicy)
	permissionsJSON, _ := json.Marshal(permissionsPolicy)
	response.IssuerURL = issuerURL
	response.TrustPolicy = trustPolicy
	response.PermissionsPolicy = permissionsPolicy
	response.Terraform = fmt.Sprintf(`resource "aws_iam_openid_connect_provider" "nuon" {
  url             = %q
  client_id_list  = ["sts.amazonaws.com"]
}

resource "aws_iam_role" "nuon_cloud_connection" {
  name               = "<role-name>"
  assume_role_policy = jsonencode(%s)
}

resource "aws_iam_role_policy" "nuon_cloud_connection" {
  name   = "nuon-cloud-connection"
  role   = aws_iam_role.nuon_cloud_connection.id
  policy = jsonencode(%s)
}`, issuerURL, policyJSON, permissionsJSON)
	response.CLI = fmt.Sprintf("aws iam create-open-id-connect-provider --url %s --client-id-list sts.amazonaws.com\naws iam create-role --role-name <role-name> --assume-role-policy-document '%s'\naws iam put-role-policy --role-name <role-name> --policy-name nuon-cloud-connection --policy-document '%s'", issuerURL, policyJSON, permissionsJSON)
	response.CloudFormation = fmt.Sprintf("NuonOIDCProvider:\n  Type: AWS::IAM::OIDCProvider\n  Properties:\n    Url: %s\n    ClientIdList:\n      - sts.amazonaws.com\nNuonConnectionRole:\n  Type: AWS::IAM::Role\n  Properties:\n    RoleName: <role-name>\n    AssumeRolePolicyDocument: %s\n    Policies:\n      - PolicyName: nuon-cloud-connection\n        PolicyDocument: %s\n", issuerURL, string(policyJSON), string(permissionsJSON))
	return response
}

func awsPermissionsPolicy(connection *app.CloudConnection, repositories []string) map[string]any {
	statements := make([]any, 0, 3)
	if connection.HasRequestedCapability(app.CloudConnectionCapabilityStacks) {
		statements = append(statements, map[string]any{
			"Sid":      "ManageNuonInstallStack",
			"Effect":   "Allow",
			"Action":   awsStackActions,
			"Resource": "*",
		})
	}
	if connection.HasRequestedCapability(app.CloudConnectionCapabilityImages) {
		statements = append(statements,
			map[string]any{"Sid": "GetECRAuthorizationToken", "Effect": "Allow", "Action": "ecr:GetAuthorizationToken", "Resource": "*"},
			map[string]any{"Sid": "AccessECRRepositories", "Effect": "Allow", "Action": awsRepositoryActions, "Resource": awsRepositoryARNs(connection, repositories)},
		)
	}
	return map[string]any{"Version": "2012-10-17", "Statement": statements}
}

func awsRepositoryARNs(connection *app.CloudConnection, repositories []string) []string {
	if len(repositories) == 0 {
		repositories = []string{"*"}
	}
	arns := make([]string, 0, len(repositories))
	for _, repository := range repositories {
		arns = append(arns, fmt.Sprintf("arn:aws:ecr:%s:%s:repository/%s", connection.DefaultRegion, connection.TargetID, strings.Trim(repository, "/")))
	}
	return arns
}

var awsStackActions = []string{
	"autoscaling:CreateAutoScalingGroup", "autoscaling:DeleteAutoScalingGroup", "autoscaling:DescribeAutoScalingGroups", "autoscaling:DescribeScalingActivities", "autoscaling:UpdateAutoScalingGroup",
	"cloudformation:CreateStack", "cloudformation:DeleteStack", "cloudformation:DescribeStackEvents", "cloudformation:DescribeStacks", "cloudformation:GetTemplate", "cloudformation:GetTemplateSummary", "cloudformation:ListStackResources", "cloudformation:UpdateStack", "cloudformation:ValidateTemplate",
	"ec2:AllocateAddress", "ec2:AssociateRouteTable", "ec2:AttachInternetGateway", "ec2:AuthorizeSecurityGroupEgress", "ec2:AuthorizeSecurityGroupIngress", "ec2:CreateInternetGateway", "ec2:CreateLaunchTemplate", "ec2:CreateNatGateway", "ec2:CreateRoute", "ec2:CreateRouteTable", "ec2:CreateSecurityGroup", "ec2:CreateSubnet", "ec2:CreateTags", "ec2:CreateVpc", "ec2:DeleteInternetGateway", "ec2:DeleteLaunchTemplate", "ec2:DeleteNatGateway", "ec2:DeleteRoute", "ec2:DeleteRouteTable", "ec2:DeleteSecurityGroup", "ec2:DeleteSubnet", "ec2:DeleteVpc", "ec2:Describe*", "ec2:DetachInternetGateway", "ec2:DisassociateRouteTable", "ec2:ModifySubnetAttribute", "ec2:ModifyVpcAttribute", "ec2:ReleaseAddress", "ec2:RevokeSecurityGroupEgress", "ec2:RevokeSecurityGroupIngress",
	"iam:AddRoleToInstanceProfile", "iam:AttachRolePolicy", "iam:CreateInstanceProfile", "iam:CreatePolicy", "iam:CreateRole", "iam:DeleteInstanceProfile", "iam:DeletePolicy", "iam:DeleteRole", "iam:DeleteRolePolicy", "iam:DetachRolePolicy", "iam:GetInstanceProfile", "iam:GetPolicy", "iam:GetPolicyVersion", "iam:GetRole", "iam:GetRolePolicy", "iam:ListAttachedRolePolicies", "iam:ListInstanceProfilesForRole", "iam:ListPolicyVersions", "iam:ListRolePolicies", "iam:PassRole", "iam:PutRolePolicy", "iam:RemoveRoleFromInstanceProfile", "iam:TagInstanceProfile", "iam:TagPolicy", "iam:TagRole", "iam:UntagInstanceProfile", "iam:UntagPolicy", "iam:UntagRole", "iam:UpdateAssumeRolePolicy",
	"lambda:AddPermission", "lambda:CreateFunction", "lambda:DeleteFunction", "lambda:GetFunction", "lambda:GetFunctionConfiguration", "lambda:RemovePermission", "lambda:TagResource", "lambda:UntagResource", "lambda:UpdateFunctionCode", "lambda:UpdateFunctionConfiguration",
	"logs:CreateLogGroup", "logs:CreateLogStream", "logs:DeleteLogGroup", "logs:DeleteLogStream", "logs:DescribeLogGroups", "logs:DescribeLogStreams", "logs:PutRetentionPolicy", "logs:TagResource", "logs:UntagResource",
	"route53resolver:AssociateFirewallRuleGroup", "route53resolver:CreateFirewallDomainList", "route53resolver:CreateFirewallRuleGroup", "route53resolver:CreateFirewallRule", "route53resolver:DeleteFirewallDomainList", "route53resolver:DeleteFirewallRuleGroup", "route53resolver:DeleteFirewallRule", "route53resolver:DisassociateFirewallRuleGroup", "route53resolver:GetFirewallConfig", "route53resolver:GetFirewallDomainList", "route53resolver:GetFirewallRuleGroup", "route53resolver:GetFirewallRuleGroupAssociation", "route53resolver:ListFirewallDomainLists", "route53resolver:ListFirewallDomains", "route53resolver:ListFirewallRuleGroupAssociations", "route53resolver:ListFirewallRuleGroups", "route53resolver:ListFirewallRules", "route53resolver:ListTagsForResource", "route53resolver:TagResource", "route53resolver:UntagResource", "route53resolver:UpdateFirewallDomains", "route53resolver:UpdateFirewallRule", "route53resolver:UpdateFirewallRuleGroupAssociation",
	"s3:GetObject",
	"secretsmanager:CreateSecret", "secretsmanager:DeleteSecret", "secretsmanager:DescribeSecret", "secretsmanager:GetResourcePolicy", "secretsmanager:PutResourcePolicy", "secretsmanager:TagResource", "secretsmanager:UntagResource", "secretsmanager:UpdateSecret",
	"iam:CreateServiceLinkedRole",
}

var awsRepositoryActions = []string{
	"ecr:BatchCheckLayerAvailability", "ecr:BatchDeleteImage", "ecr:BatchGetImage", "ecr:CompleteLayerUpload", "ecr:DescribeImages", "ecr:DescribeRepositories", "ecr:GetDownloadUrlForLayer", "ecr:InitiateLayerUpload", "ecr:ListImages", "ecr:ListTagsForResource", "ecr:PutImage", "ecr:UploadLayerPart",
}

func (s *service) azureSetup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	if options.Repositories == nil {
		options.Repositories = []string{}
	}
	response := SetupResponse{Subject: subject(connection), Audience: azureTokenExchangeAudience, Capabilities: connection.RequestedCapabilities, Repositories: options.Repositories, Registry: options.Registry}
	if s.issuer == nil {
		return response
	}
	issuerURL := s.issuer.Issuer()
	federatedCredential := map[string]any{
		"name":      "nuon-cloud-connection",
		"issuer":    issuerURL,
		"subject":   subject(connection),
		"audiences": []string{azureTokenExchangeAudience},
	}
	portalJSON, _ := json.MarshalIndent(federatedCredential, "", "  ")
	capabilitiesJSON, _ := json.Marshal(connection.RequestedCapabilities)
	repositoriesJSON, _ := json.Marshal(options.Repositories)
	response.IssuerURL = issuerURL
	response.TrustPolicy = federatedCredential
	response.PortalJSON = string(portalJSON)
	response.Terraform = fmt.Sprintf(`module "nuon_cloud_connection" {
  source          = "nuonco/acr-access/azure"
  nuon_issuer     = %q
  nuon_subject    = %q
  capabilities    = %s
  repositories    = %s
  subscription_id = %q
  registry        = %q
}`, issuerURL, subject(connection), capabilitiesJSON, repositoriesJSON, connection.TargetID, options.Registry)
	lines := []string{
		"APP_ID=$(az ad app create --display-name nuon-cloud-connection --query appId -o tsv)",
		"cat > nuon-federated-credential.json <<'JSON'",
		string(portalJSON),
		"JSON",
		"az ad app federated-credential create --id \"$APP_ID\" --parameters nuon-federated-credential.json",
	}
	if connection.HasRequestedCapability(app.CloudConnectionCapabilityStacks) {
		lines = append(lines, fmt.Sprintf("az role assignment create --assignee \"$APP_ID\" --role Owner --scope /subscriptions/%s", connection.TargetID))
	}
	if connection.HasRequestedCapability(app.CloudConnectionCapabilityImages) {
		lines = append(lines, "REGISTRY_ID=$(az acr show --name "+options.Registry+" --query id -o tsv)", "az role assignment create --assignee \"$APP_ID\" --role AcrPull --scope \"$REGISTRY_ID\"")
	}
	response.CLI = strings.Join(lines, "\n")
	return response
}

func (s *service) gcpSetup(connection *app.CloudConnection, options SetupOptions) SetupResponse {
	if options.Repositories == nil {
		options.Repositories = []string{}
	}
	provider := connection.IdentityProvider
	if provider == "" {
		provider = "projects/<project-number>/locations/global/workloadIdentityPools/<pool-id>/providers/<provider-id>"
	}
	audience := "https://iam.googleapis.com/" + strings.TrimPrefix(strings.TrimPrefix(provider, "https://iam.googleapis.com/"), "//iam.googleapis.com/")
	response := SetupResponse{Subject: subject(connection), Audience: audience, Capabilities: connection.RequestedCapabilities, Repositories: options.Repositories}
	if s.issuer == nil {
		return response
	}
	issuerURL := s.issuer.Issuer()
	serviceAccountID := strings.SplitN(connection.Principal, "@", 2)[0]
	capabilitiesJSON, _ := json.Marshal(connection.RequestedCapabilities)
	repositoriesJSON, _ := json.Marshal(options.Repositories)
	response.IssuerURL = issuerURL
	response.Terraform = fmt.Sprintf(`module "nuon_cloud_connection" {
  source       = "nuonco/gar-access/google"
  nuon_issuer  = %q
  nuon_subject = %q
  capabilities = %s
  repositories = %s
  project_id   = %q
}`, issuerURL, subject(connection), capabilitiesJSON, repositoriesJSON, connection.TargetID)
	lines := []string{
		"PROJECT_ID=" + connection.TargetID,
		"PROJECT_NUMBER=$(gcloud projects describe \"$PROJECT_ID\" --format='value(projectNumber)')",
		"POOL_ID=<pool-id>",
		"PROVIDER_ID=<provider-id>",
		"SERVICE_ACCOUNT_EMAIL=" + connection.Principal,
		"gcloud iam workload-identity-pools create \"$POOL_ID\" --project \"$PROJECT_ID\" --location global --display-name \"Nuon cloud connection\"",
		fmt.Sprintf("gcloud iam workload-identity-pools providers create-oidc \"$PROVIDER_ID\" --project \"$PROJECT_ID\" --location global --workload-identity-pool \"$POOL_ID\" --issuer-uri %q --allowed-audiences %q --attribute-mapping=\"google.subject=assertion.sub\"", issuerURL, audience),
		"gcloud iam service-accounts create " + serviceAccountID + " --project \"$PROJECT_ID\"",
		fmt.Sprintf("gcloud iam service-accounts add-iam-policy-binding \"$SERVICE_ACCOUNT_EMAIL\" --project \"$PROJECT_ID\" --role roles/iam.workloadIdentityUser --member \"principal://iam.googleapis.com/projects/$PROJECT_NUMBER/locations/global/workloadIdentityPools/$POOL_ID/subject/%s\"", subject(connection)),
	}
	if len(options.Repositories) == 0 {
		lines = append(lines, "gcloud projects add-iam-policy-binding \"$PROJECT_ID\" --role roles/artifactregistry.reader --member \"serviceAccount:$SERVICE_ACCOUNT_EMAIL\"")
	} else {
		for _, repository := range options.Repositories {
			parts := strings.SplitN(strings.Trim(repository, "/"), "/", 2)
			location, name := "<location>", repository
			if len(parts) == 2 {
				location, name = parts[0], parts[1]
			}
			lines = append(lines, fmt.Sprintf("gcloud artifacts repositories add-iam-policy-binding %q --project \"$PROJECT_ID\" --location %q --role roles/artifactregistry.reader --member \"serviceAccount:$SERVICE_ACCOUNT_EMAIL\"", name, location))
		}
	}
	lines = append(lines, "gcloud iam workload-identity-pools providers describe \"$PROVIDER_ID\" --project \"$PROJECT_ID\" --location global --workload-identity-pool \"$POOL_ID\" --format='value(name)'", "Paste the provider resource name returned above into identity_provider.")
	response.CLI = strings.Join(lines, "\n")
	return response
}
