export const ACCOUNT_ID = '123456789012'
export const SUBJECT = 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE'

const CLI = [
  `aws iam create-open-id-connect-provider \\
  --url https://api.nuon.co \\
  --client-id-list sts.amazonaws.com`,
  `aws iam create-role \\
  --role-name nuon-cloud-connection \\
  --assume-role-policy-document file://nuon-trust.json

# nuon-trust.json fixes the audience to sts.amazonaws.com
# and subject to ${SUBJECT}`,
  `aws iam put-role-policy \\
  --role-name nuon-cloud-connection \\
  --policy-name NuonConnectionPermissions \\
  --policy-document file://nuon-permissions.json

# nuon-permissions.json scopes ECR reads to:
# acme/api and acme/worker in us-east-1
# Install stack names must begin with nuon-`,
  `aws iam get-role --role-name nuon-cloud-connection \\
  --query Role.Arn --output text

# Paste: arn:aws:iam::${ACCOUNT_ID}:role/nuon-cloud-connection`,
  `# Nuon runs these checks after you select Verify connection:
# 1. Exchange OIDC token for role credentials
# 2. Prove a foreign org subject is denied
# 3. Probe CloudFormation, IAM, and ECR permissions`,
]

const TERRAFORM = [
  `resource "aws_iam_openid_connect_provider" "nuon" {
  url             = "https://api.nuon.co"
  client_id_list  = ["sts.amazonaws.com"]
}`,
  `module "nuon_connection_role" {
  source        = "nuonco/ecr-access/aws"
  account_id    = "${ACCOUNT_ID}"
  oidc_subject  = "${SUBJECT}"
}`,
  `module "nuon_connection_role" {
  repositories      = ["acme/api", "acme/worker"]
  region            = "us-east-1"
  stack_name_prefix = "nuon-"
  capabilities      = ["stacks", "images"]
}`,
  `output "nuon_role_arn" {
  value = module.nuon_connection_role.role_arn
}`,
  `# In Nuon, paste output.nuon_role_arn and select Verify connection.
# Nuon tests the exact and foreign OIDC subjects, then each capability.`,
]

const CLOUDFORMATION = [
  `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  NuonOIDCProvider:
    Type: AWS::IAM::OIDCProvider
    Properties:
      Url: https://api.nuon.co
      ClientIdList: [sts.amazonaws.com]`,
  `Resources:
  NuonConnectionRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: nuon-cloud-connection
      AssumeRolePolicyDocument:
        Statement:
          - Action: sts:AssumeRoleWithWebIdentity
            Condition:
              StringEquals:
                api.nuon.co:sub: ${SUBJECT}`,
  `Resources:
  NuonConnectionRole:
    Properties:
      Policies:
        - PolicyName: NuonConnectionPermissions
          PolicyDocument:
            # CloudFormation/IAM for nuon-* stacks
            # ECR reads for acme/api and acme/worker`,
  `Outputs:
  NuonRoleArn:
    Value: !GetAtt NuonConnectionRole.Arn`,
  `# Open the quick-create stack, then paste the NuonRoleArn output.
# Nuon verifies OIDC isolation and selected capabilities.`,
]

export const snippets = {
  cli: CLI,
  terraform: TERRAFORM,
  cloudformation: CLOUDFORMATION,
} as const

export const POLICY = `{
  "Statement": [
    { "Effect": "Allow", "Action": ["cloudformation:CreateStack", "cloudformation:UpdateStack", "cloudformation:DeleteStack", "cloudformation:DescribeStacks", "cloudformation:DescribeStackEvents", "iam:CreateRole", "iam:AttachRolePolicy", "iam:PassRole"], "Resource": "arn:aws:cloudformation:*:${ACCOUNT_ID}:stack/nuon-*/*" },
    { "Effect": "Allow", "Action": ["ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer", "ecr:BatchCheckLayerAvailability", "ecr:DescribeRepositories", "ecr:DescribeImages"], "Resource": ["arn:aws:ecr:us-east-1:${ACCOUNT_ID}:repository/acme/api", "arn:aws:ecr:us-east-1:${ACCOUNT_ID}:repository/acme/worker"] },
    { "Effect": "Allow", "Action": "ecr:GetAuthorizationToken", "Resource": "*" }
  ]
}`
