export const ACCOUNT_ID = '123456789012'
export const SUBJECT = 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE'

export const CAPABILITIES = {
  stacks: {
    title: 'Install stacks',
    what: 'Create, update, and delete the CloudFormation stack that bootstraps each install.',
    why: 'Customers do not have to create runner roles, networking, or other install resources by hand.',
    actions:
      'cloudformation:CreateStack · UpdateStack · DeleteStack · DescribeStacks · DescribeStackEvents · iam:CreateRole · AttachRolePolicy · PassRole',
  },
  images: {
    title: 'Pull images',
    what: 'Read private images from ECR for Nuon builds.',
    why: 'The build runner needs source images to publish them into each install registry.',
    actions:
      'ecr:GetAuthorizationToken · BatchGetImage · GetDownloadUrlForLayer · BatchCheckLayerAvailability · DescribeRepositories · DescribeImages',
  },
} as const

export const scriptFor = (
  format: string,
  stacks: boolean,
  images: boolean,
  repos: string[],
  region: string,
  prefix: string
) => {
  const scopes = `${stacks ? `stack_name_prefix = "${prefix}"` : '# Install stacks disabled'}\n  ${images ? `repositories = [${repos.map((repo) => `"${repo}"`).join(', ')}]\n  region = "${region}"` : '# Pull images disabled'}`
  if (format === 'cli')
    return `aws iam create-open-id-connect-provider --url https://api.nuon.co --client-id-list sts.amazonaws.com

aws iam create-role --role-name nuon-cloud-connection \\
  --assume-role-policy-document file://trust-policy.json

aws iam put-role-policy --role-name nuon-cloud-connection \\
  --policy-name NuonConnectionPermissions \\
  --policy-document file://permissions-policy.json

# Selected scope: ${stacks ? `stacks ${prefix}` : 'no stacks'}; ${images ? `${repos.join(', ')} in ${region}` : 'no ECR access'}
aws iam get-role --role-name nuon-cloud-connection --query Role.Arn --output text`
  if (format === 'cloudformation')
    return `AWSTemplateFormatVersion: '2010-09-09'
Resources:
  NuonConnectionRole:
    Type: AWS::IAM::Role
    Properties:
      RoleName: nuon-cloud-connection
      AssumeRolePolicyDocument:
        Statement:
          - Effect: Allow
            Action: sts:AssumeRoleWithWebIdentity
            Condition:
              StringEquals:
                api.nuon.co:sub: ${SUBJECT}
                api.nuon.co:aud: sts.amazonaws.com
      # Scope: ${stacks ? prefix : 'no stacks'} / ${images ? repos.join(', ') : 'no images'}`
  return `module "nuon_cloud_connection" {
  source = "nuonco/ecr-access/aws"

  account_id    = "${ACCOUNT_ID}"
  oidc_issuer   = "https://api.nuon.co"
  oidc_subject  = "${SUBJECT}"
  oidc_audience = "sts.amazonaws.com"
  ${scopes}
}`
}

export const CUSTOM_POLICY = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["ecr:GetAuthorizationToken", "ecr:BatchGetImage", "ecr:GetDownloadUrlForLayer"],
    "Resource": ["arn:aws:ecr:us-east-1:123456789012:repository/acme/api", "arn:aws:ecr:us-east-1:123456789012:repository/acme/worker"]
  }]
}`
