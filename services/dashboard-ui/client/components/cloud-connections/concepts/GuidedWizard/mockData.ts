export const ACCOUNT_ID = '123456789012'
export const SUBJECT = 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE'

export const STACK_ACTIONS = [
  [
    'Create, update, delete, and inspect install stacks',
    'cloudformation:CreateStack · UpdateStack · DeleteStack · DescribeStacks · DescribeStackEvents',
  ],
  [
    'Create runner roles and attach their policies',
    'iam:CreateRole · AttachRolePolicy · PassRole',
  ],
] as const

export const IMAGE_ACTIONS = [
  [
    'Get a temporary ECR login token',
    'ecr:GetAuthorizationToken (account-wide by AWS design)',
  ],
  [
    'Read image manifests and layers from the selected repositories',
    'ecr:BatchGetImage · GetDownloadUrlForLayer · BatchCheckLayerAvailability · DescribeRepositories · DescribeImages',
  ],
] as const

export const TRUST_POLICY = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": { "Federated": "arn:aws:iam::123456789012:oidc-provider/api.nuon.co" },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "api.nuon.co:aud": "sts.amazonaws.com",
        "api.nuon.co:sub": "org:org_01JEXAMPLE:connection:cc_01JEXAMPLE"
      }
    }
  }]
}`

export const permissionsPolicy = (
  images = true,
  stackPrefix = 'nuon-*',
  repos = '*',
  region = 'us-east-1'
) => `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "InstallStacks",
      "Effect": "Allow",
      "Action": ["cloudformation:CreateStack", "cloudformation:UpdateStack", "cloudformation:DeleteStack", "cloudformation:DescribeStacks", "cloudformation:DescribeStackEvents", "iam:CreateRole", "iam:AttachRolePolicy", "iam:PassRole"],
      "Resource": "arn:aws:cloudformation:${region}:${ACCOUNT_ID}:stack/${stackPrefix}/*"
    }${
      images
        ? `,
    {
      "Sid": "PullImages",
      "Effect": "Allow",
      "Action": ["ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer", "ecr:BatchCheckLayerAvailability", "ecr:DescribeRepositories", "ecr:DescribeImages"],
      "Resource": "arn:aws:ecr:${region}:${ACCOUNT_ID}:repository/${repos}"
    },
    { "Effect": "Allow", "Action": "ecr:GetAuthorizationToken", "Resource": "*" }`
        : ''
    }
  ]
}`

export const terraformSnippet = (
  repos = '*',
  region = 'us-east-1',
  stackPrefix = 'nuon-*'
) => `module "nuon_cloud_connection" {
  source  = "nuonco/ecr-access/aws"

  account_id       = "${ACCOUNT_ID}"
  region           = "${region}"
  oidc_issuer      = "https://api.nuon.co"
  oidc_subject     = "${SUBJECT}"
  oidc_audience    = "sts.amazonaws.com"
  repositories     = ["${repos}"]
  stack_name_prefix = "${stackPrefix}"
}`

export const cliSnippet = (
  region = 'us-east-1'
) => `aws iam create-open-id-connect-provider \\
  --url https://api.nuon.co \\
  --client-id-list sts.amazonaws.com

aws iam create-role \\
  --role-name nuon-cloud-connection \\
  --assume-role-policy-document file://nuon-trust.json

aws iam put-role-policy \\
  --role-name nuon-cloud-connection \\
  --policy-name NuonConnectionPermissions \\
  --policy-document file://nuon-permissions.json

aws iam get-role --role-name nuon-cloud-connection \\
  --query Role.Arn --output text --region ${region}`

export const cloudFormationSnippet = (
  stackPrefix = 'nuon-*'
) => `AWSTemplateFormatVersion: '2010-09-09'
Parameters:
  NuonSubject:
    Type: String
    Default: ${SUBJECT}
Resources:
  NuonConnectionRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: nuon-cloud-connection
      AssumeRolePolicyDocument: ${TRUST_POLICY.replaceAll('\n', '\n        ')}
      Policies:
        - PolicyName: NuonConnectionPermissions
          PolicyDocument:
            Statement:
              - Effect: Allow
                Action: [cloudformation:CreateStack, cloudformation:UpdateStack, cloudformation:DeleteStack]
                Resource: arn:aws:cloudformation:*:${ACCOUNT_ID}:stack/${stackPrefix}/*`
