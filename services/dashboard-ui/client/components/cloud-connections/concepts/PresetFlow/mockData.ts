export const ACCOUNT_ID = '123456789012'
export const REGION = 'us-east-1'
export const SUBJECT = 'org:org_01JEXAMPLE:connection:cc_01JEXAMPLE'
export const ROLE_ARN = `arn:aws:iam::${ACCOUNT_ID}:role/nuon-cloud-connection`

export const STACK_ACTIONS = [
  {
    purpose: 'Create and change the install stack',
    actions: [
      'cloudformation:CreateStack',
      'cloudformation:UpdateStack',
      'cloudformation:DeleteStack',
    ],
  },
  {
    purpose: 'Read stack status, events, resources, templates, and outputs',
    actions: [
      'cloudformation:DescribeStacks',
      'cloudformation:DescribeStackEvents',
      'cloudformation:DescribeStackResources',
      'cloudformation:GetTemplate',
    ],
  },
] as const

export const STACK_RESOURCE = `arn:aws:cloudformation:${REGION}:${ACCOUNT_ID}:stack/nuon-*/*`

export const TRUST_POLICY = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": {
      "Federated": "arn:aws:iam::${ACCOUNT_ID}:oidc-provider/api.nuon.co"
    },
    "Action": "sts:AssumeRoleWithWebIdentity",
    "Condition": {
      "StringEquals": {
        "api.nuon.co:aud": "sts.amazonaws.com",
        "api.nuon.co:sub": "${SUBJECT}"
      }
    }
  }]
}`

export const PERMISSIONS_POLICY = `{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "ManageNuonInstallStacks",
    "Effect": "Allow",
    "Action": [
      "cloudformation:CreateStack",
      "cloudformation:UpdateStack",
      "cloudformation:DeleteStack",
      "cloudformation:DescribeStacks",
      "cloudformation:DescribeStackEvents",
      "cloudformation:DescribeStackResources",
      "cloudformation:GetTemplate"
    ],
    "Resource": "${STACK_RESOURCE}"
  }]
}`

type TFormat = 'terraform' | 'cli' | 'cloudformation'

const snippets: Record<
  TFormat,
  { oidc: string; role: string; permissions: string }
> = {
  terraform: {
    oidc: `resource "aws_iam_openid_connect_provider" "nuon" {
  url            = "https://api.nuon.co"
  client_id_list = ["sts.amazonaws.com"]
}`,
    role: `resource "aws_iam_role" "nuon_connection" {
  name               = "nuon-cloud-connection"
  assume_role_policy = file("\${path.module}/nuon-trust.json")
}`,
    permissions: `resource "aws_iam_role_policy" "nuon_stacks" {
  name   = "NuonStacks"
  role   = aws_iam_role.nuon_connection.id
  policy = file("\${path.module}/nuon-permissions.json")
}`,
  },
  cli: {
    oidc: `aws iam create-open-id-connect-provider \\
  --url https://api.nuon.co \\
  --client-id-list sts.amazonaws.com`,
    role: `aws iam create-role \\
  --role-name nuon-cloud-connection \\
  --assume-role-policy-document file://nuon-trust.json`,
    permissions: `aws iam put-role-policy \\
  --role-name nuon-cloud-connection \\
  --policy-name NuonStacks \\
  --policy-document file://nuon-permissions.json`,
  },
  cloudformation: {
    oidc: `NuonOIDCProvider:
  Type: AWS::IAM::OIDCProvider
  Properties:
    Url: https://api.nuon.co
    ClientIdList: [sts.amazonaws.com]`,
    role: `NuonConnectionRole:
  Type: AWS::IAM::Role
  Properties:
    RoleName: nuon-cloud-connection
    AssumeRolePolicyDocument: # Use the trust policy from Nuon`,
    permissions: `NuonConnectionRole:
  Properties:
    Policies:
      - PolicyName: NuonStacks
        PolicyDocument: # Use the Stacks policy from Nuon`,
  },
}

export const setupSnippet = (
  format: TFormat,
  part: keyof (typeof snippets)[TFormat]
) => {
  if (part !== 'role') return snippets[format][part]
  if (format === 'cli')
    return `cat > nuon-trust.json <<'POLICY'
${TRUST_POLICY}
POLICY

aws iam create-role \\
  --role-name nuon-cloud-connection \\
  --assume-role-policy-document file://nuon-trust.json`
  if (format === 'terraform')
    return `resource "aws_iam_role" "nuon_connection" {
  name = "nuon-cloud-connection"
  assume_role_policy = <<POLICY
${TRUST_POLICY}
POLICY
}`
  return `NuonConnectionRole:
  Type: AWS::IAM::Role
  Properties:
    RoleName: nuon-cloud-connection
    AssumeRolePolicyDocument: ${TRUST_POLICY.replaceAll('\n', '\n      ')}`
}
