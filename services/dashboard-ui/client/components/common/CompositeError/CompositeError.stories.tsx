import { CompositeError } from './CompositeError'
import type { TCompositeError } from '@/types'

export default {
  title: 'Common/CompositeError',
}

const awsPermissionError: TCompositeError = {
  type: 'aws_permission_error',
  severity: 'error',
  message: 'Deploy failed because the install role is missing required AWS IAM permissions.',
  sections: [
    {
      heading: 'Missing permissions',
      body: '- `ec2:CreateSecurityGroup`\n- `ec2:AuthorizeSecurityGroupIngress`',
    },
    {
      heading: 'How to fix',
      body: 'Add the missing actions to the install role policy, then retry the deploy.',
    },
  ],
}

export const Default = () => <CompositeError error={awsPermissionError} />

export const NoSections = () => (
  <CompositeError
    error={{
      type: 'aws_permission_error',
      severity: 'error',
      message: 'The install role is not authorized to perform this Terraform operation.',
    }}
  />
)

export const Warning = () => (
  <CompositeError
    error={{
      type: 'aws_permission_error',
      severity: 'warning',
      message: 'Some optional permissions are missing and may limit functionality.',
      sections: [{ heading: 'Details', body: 'Missing `s3:GetBucketTagging`.' }],
    }}
  />
)

const componentBuildUnavailable: TCompositeError = {
  type: 'deploy.component_build_unavailable',
  severity: 'error',
  message: 'Build for rds_cluster_coder failed',
  sections: [
    {
      heading: 'Why',
      body: "Deploying rds_cluster_coder needs a build that completed successfully. Its most recent build failed, so the deploy can't continue.",
    },
    {
      heading: 'Build status',
      body: 'Build: `bld19ycmcr7zzjecyvetn22v30`\n\nStatus: `error`\n\n```\nbuild failed: no heart beats found\n```',
    },
    {
      heading: 'How to fix',
      body: 'Fix what caused the build to fail, then rebuild rds_cluster_coder with the latest config. Once the build is active, retry the deploy.',
    },
  ],
}

export const ComponentBuildUnavailable = () => (
  <CompositeError error={componentBuildUnavailable} />
)

const terraformPlanFailed: TCompositeError = {
  type: 'deploy.terraform_plan_failed',
  severity: 'error',
  message: 'Terraform plan failed for vpc_networking',
  sections: [
    {
      heading: 'Why',
      body: 'The plan step exited with an error before any changes were applied. No resources were created, modified, or destroyed.',
    },
    {
      heading: 'Error output',
      body: '```\nError: creating EC2 Subnet: InvalidSubnet.Range:\nThe CIDR \'10.0.1.0/24\' conflicts with another subnet\n\n  with module.networking.aws_subnet.private[1],\n  on networking/subnets.tf line 14, in resource "aws_subnet" "private":\n  14: resource "aws_subnet" "private" {\n```',
    },
    {
      heading: 'Affected resources',
      body: '| Resource | Action | Status |\n| --- | --- | --- |\n| `aws_subnet.private[0]` | create | ok |\n| `aws_subnet.private[1]` | create | failed |\n| `aws_route_table.private` | create | skipped |',
    },
    {
      heading: 'How to fix',
      body: 'Update the `private_subnet_cidrs` input so each subnet has a non-overlapping range, then retry the deploy. See the [networking guide](https://docs.nuon.co/networking) for recommended CIDR layouts.',
    },
  ],
}

export const TerraformPlanFailed = () => (
  <CompositeError error={terraformPlanFailed} />
)

const runnerHealthFatal: TCompositeError = {
  type: 'runner.unreachable',
  severity: 'fatal',
  message: 'Runner for install acme-prod is unreachable',
  sections: [
    {
      heading: 'Why',
      body: "The control plane hasn't received a heartbeat from this install's runner in over **15 minutes**. Deploys, builds, and actions cannot be scheduled until the runner reconnects.",
    },
    {
      heading: 'Last seen',
      body: 'Runner: `rnr8k2jdh4qp1xz9`\n\nLast heartbeat: `2026-06-29T14:02:11Z`\n\nRegion: `us-west-2`',
    },
    {
      heading: 'Common causes',
      body: '1. The runner pod was evicted or `OOMKilled`\n2. Outbound network egress to `*.nuon.co` is blocked\n3. The runner service account lost permission to assume its IAM role',
    },
    {
      heading: 'How to fix',
      body: 'Check the runner pod status in the customer cluster:\n\n```\nkubectl get pods -n nuon-runner\nkubectl logs -n nuon-runner deploy/nuon-runner --tail=100\n```\n\nIf the pod is healthy, verify egress and IAM role trust, then wait for the next heartbeat.',
    },
  ],
}

export const RunnerUnreachable = () => (
  <CompositeError error={runnerHealthFatal} />
)

const policyViolationInfo: TCompositeError = {
  type: 'policy.guardrail_warning',
  severity: 'info',
  message: 'Deploy passed with policy advisories',
  sections: [
    {
      heading: 'Why',
      body: 'The deploy completed, but the policy engine flagged a few advisory checks that did not block the rollout.',
    },
    {
      heading: 'Advisories',
      body: '- `cost`: estimated monthly increase of **$420** from the new `db.r6g.2xlarge` instance\n- `tagging`: `aws_db_instance.main` is missing the `owner` tag\n- `encryption`: storage encryption uses the default AWS-managed key, not a customer-managed key',
    },
    {
      heading: 'How to fix',
      body: 'These are non-blocking. To clear them, right-size the instance, add the missing tag, and switch to a CMK in the next config revision.',
    },
  ],
}

export const PolicyAdvisories = () => (
  <CompositeError error={policyViolationInfo} />
)
