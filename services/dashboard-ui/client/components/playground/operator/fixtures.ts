import type { TCompositeError, TInstall } from '@/types'

export type TOperatorError = TCompositeError & {
  action?: { label: string; href: string }
}

export type TOperatorInstall = TInstall & {
  app_config_version?: number
  app_config_latest_version?: number
  deployments_status?: string
  deployments_detail?: string
  resources_status?: string
  resources_detail?: string
  health_status?: string
  health_detail?: string
  errors?: TOperatorError[]
}

const ORG_ID = 'org-operator-demo'

export const operatorInstalls = [
  {
    id: 'insta1b2c3d4e5f6g7h8i9j0k1',
    org_id: ORG_ID,
    app_id: 'app-acme-platform',
    name: 'acme-production',
    app: { id: 'app-acme-platform', name: 'acme-platform' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 42,
    app_config_latest_version: 42,
    deployments_status: 'active',
    deployments_detail: '6 of 6 components deployed',
    resources_status: 'ok',
    resources_detail: '38 resources, no drift',
    health_status: 'healthy',
    health_detail: 'All components passing health checks',
    cloud_platform: 'aws',
    aws_account: { region: 'us-east-1' },
    labels: { tier: 'enterprise', env: 'prod' },
    updated_at: '2026-09-22T09:12:00Z',
    created_at: '2026-01-14T10:00:00Z',
  },
  {
    id: 'instb2c3d4e5f6g7h8i9j0k1l2',
    org_id: ORG_ID,
    app_id: 'app-acme-platform',
    name: 'globex-production',
    app: { id: 'app-acme-platform', name: 'acme-platform' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 38,
    app_config_latest_version: 42,
    deployments_status: 'pending-approval',
    deployments_detail: 'apply_changes is waiting on an approver',
    resources_status: 'ok',
    resources_detail: '41 resources, no drift',
    health_status: 'healthy',
    health_detail: 'All components passing health checks',
    cloud_platform: 'aws',
    aws_account: { region: 'eu-west-1' },
    labels: { tier: 'enterprise', env: 'prod' },
    updated_at: '2026-08-30T16:40:00Z',
    created_at: '2026-02-02T11:20:00Z',
  },
  {
    id: 'instc3d4e5f6g7h8i9j0k1l2m3',
    org_id: ORG_ID,
    app_id: 'app-acme-platform',
    name: 'initech-production',
    app: { id: 'app-acme-platform', name: 'acme-platform' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 31,
    app_config_latest_version: 42,
    deployments_status: 'active',
    deployments_detail: '6 of 6 components deployed',
    resources_status: 'ok',
    resources_detail: '35 resources, no drift',
    health_status: 'degraded',
    health_detail: 'api-gateway has 1 of 3 replicas available',
    errors: [
      {
        type: 'component_health_degraded',
        severity: 'warning',
        message:
          'api-gateway has 1 of 3 replicas available and has been degraded for 6 days.',
        sections: [
          {
            heading: 'Failing pods',
            kind: 'code',
            body: 'api-gateway-7d9c4f88b5-2xk9p   0/1   ImagePullBackOff\napi-gateway-7d9c4f88b5-n4rtz   0/1   ImagePullBackOff',
          },
          {
            heading: 'Likely cause',
            body: 'The image tag referenced by config v31 was removed from the registry. Updating to v42 pins a tag that still exists.',
          },
        ],
        action: {
          label: 'View api-gateway component',
          href: `/${ORG_ID}/installs/instc3d4e5f6g7h8i9j0k1l2m3/components`,
        },
      },
    ],
    cloud_platform: 'azure',
    azure_account: { location: 'eastus' },
    labels: { tier: 'growth', env: 'prod' },
    updated_at: '2026-07-11T08:05:00Z',
    created_at: '2026-03-19T09:45:00Z',
  },
  {
    id: 'instd4e5f6g7h8i9j0k1l2m3n4',
    org_id: ORG_ID,
    app_id: 'app-acme-platform',
    name: 'umbrella-production',
    app: { id: 'app-acme-platform', name: 'acme-platform' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 42,
    app_config_latest_version: 42,
    deployments_status: 'failed',
    deployments_detail: 'worker failed to deploy 8 hours ago',
    resources_status: 'ok',
    resources_detail: '44 resources, no drift',
    health_status: 'unhealthy',
    health_detail: 'worker deployment is in CrashLoopBackOff',
    errors: [
      {
        type: 'terraform_apply_failed',
        severity: 'fatal',
        message:
          'Deploy of component "worker" failed during terraform apply 8 hours ago.',
        sections: [
          {
            heading: 'Terraform output',
            kind: 'code',
            body: 'Error: creating ECS Service (acme-worker): InvalidParameterException:\n  The target group with targetGroupArn arn:aws:elasticloadbalancing:us-west-2:…:targetgroup/acme-worker\n  does not have an associated load balancer.',
          },
          {
            heading: 'How to fix',
            body: 'The load balancer was deleted outside of Nuon. Reprovision the sandbox, then redeploy.',
          },
        ],
        action: {
          label: 'Open failed deploy run',
          href: `/${ORG_ID}/installs/instd4e5f6g7h8i9j0k1l2m3n4/history/wkfl-deploy-failed`,
        },
      },
      {
        type: 'component_health_unhealthy',
        severity: 'error',
        message: 'worker deployment is in CrashLoopBackOff (412 restarts).',
        sections: [
          {
            heading: 'Last container log',
            kind: 'code',
            body: 'panic: dial tcp 10.0.3.14:5432: connect: connection refused\n\ngoroutine 1 [running]:\nmain.main()',
          },
        ],
        action: {
          label: 'View worker logs',
          href: `/${ORG_ID}/installs/instd4e5f6g7h8i9j0k1l2m3n4/components`,
        },
      },
    ],
    cloud_platform: 'aws',
    aws_account: { region: 'us-west-2' },
    labels: { tier: 'enterprise', env: 'prod' },
    updated_at: '2026-09-22T07:51:00Z',
    created_at: '2025-11-06T14:30:00Z',
  },
  {
    id: 'inste5f6g7h8i9j0k1l2m3n4o5',
    org_id: ORG_ID,
    app_id: 'app-acme-edge',
    name: 'hooli-production',
    app: { id: 'app-acme-edge', name: 'acme-edge' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 12,
    app_config_latest_version: 14,
    deployments_status: 'active',
    deployments_detail: '4 of 4 components deployed',
    resources_status: 'drifted',
    resources_detail: '1 of 22 resources drifted (edge-proxy)',
    health_status: 'healthy',
    health_detail: 'All components passing health checks',
    errors: [
      {
        type: 'resource_drift',
        severity: 'warning',
        message:
          'google_compute_backend_service.edge_proxy has drifted from the config in v12.',
        sections: [
          {
            heading: 'Drifted attributes',
            kind: 'code',
            body: '~ timeout_sec:        30 -> 120\n~ connection_draining_timeout_sec: 300 -> 60',
          },
          {
            heading: 'How to fix',
            body: 'Someone changed these in the GCP console. Redeploy to restore the config values, or fold the change into the app config.',
          },
        ],
        action: {
          label: 'Review drift',
          href: `/${ORG_ID}/installs/inste5f6g7h8i9j0k1l2m3n4o5/drift`,
        },
      },
    ],
    cloud_platform: 'gcp',
    gcp_account: { region: 'us-central1' },
    labels: { tier: 'growth', env: 'prod' },
    updated_at: '2026-09-18T12:00:00Z',
    created_at: '2026-04-22T16:10:00Z',
  },
  {
    id: 'instf6g7h8i9j0k1l2m3n4o5p6',
    org_id: ORG_ID,
    app_id: 'app-acme-edge',
    name: 'soylent-staging',
    app: { id: 'app-acme-edge', name: 'acme-edge' },
    app_branch: { id: 'brnch-staging', name: 'staging' },
    app_config_version: 14,
    app_config_latest_version: 14,
    deployments_status: 'deploying',
    deployments_detail: 'Deploying 2 of 4 components',
    resources_status: 'ok',
    resources_detail: '19 resources, no drift',
    health_status: 'healthy',
    health_detail: 'All components passing health checks',
    cloud_platform: 'aws',
    aws_account: { region: 'us-east-2' },
    labels: { tier: 'trial', env: 'staging' },
    updated_at: '2026-09-22T11:58:00Z',
    created_at: '2026-09-01T13:15:00Z',
  },
  {
    id: 'instg7h8i9j0k1l2m3n4o5p6q7',
    org_id: ORG_ID,
    app_id: 'app-acme-edge',
    name: 'vehement-production',
    app: { id: 'app-acme-edge', name: 'acme-edge' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 9,
    app_config_latest_version: 14,
    deployments_status: 'active',
    deployments_detail: '4 of 4 components deployed',
    resources_status: 'drifted',
    resources_detail: '3 of 17 resources drifted',
    health_status: 'unknown',
    health_detail: 'Runner has not reported since 2 Jun',
    errors: [
      {
        type: 'runner_unreachable',
        severity: 'error',
        message:
          'No heartbeat from this install runner in 112 days. Health and drift are last-known values, not live.',
        sections: [
          {
            heading: 'What this means',
            body: 'Everything shown for this install is stale. It may be healthy, broken, or already torn down — Nuon cannot tell.',
          },
          {
            heading: 'How to fix',
            body: 'Check the runner task in the customer account, then reprovision it if the task is gone.',
          },
        ],
        action: {
          label: 'Open install runner',
          href: `/${ORG_ID}/installs/instg7h8i9j0k1l2m3n4o5p6q7/runner`,
        },
      },
      {
        type: 'resource_drift',
        severity: 'warning',
        message: '3 of 17 resources drifted before the runner went quiet.',
        action: {
          label: 'Review drift',
          href: `/${ORG_ID}/installs/instg7h8i9j0k1l2m3n4o5p6q7/drift`,
        },
      },
    ],
    cloud_platform: 'aws',
    aws_account: { region: 'ap-southeast-2' },
    labels: { tier: 'growth', env: 'prod' },
    updated_at: '2026-06-02T19:25:00Z',
    created_at: '2025-12-11T08:00:00Z',
  },
  {
    id: 'insth8i9j0k1l2m3n4o5p6q7r8',
    org_id: ORG_ID,
    app_id: 'app-acme-platform',
    name: 'cyberdyne-sandbox',
    app: { id: 'app-acme-platform', name: 'acme-platform' },
    app_branch: { id: 'brnch-main', name: 'main' },
    app_config_version: 42,
    app_config_latest_version: 42,
    deployments_status: 'queued',
    deployments_detail: 'Waiting on sandbox provisioning',
    resources_status: 'provisioning',
    resources_detail: 'Sandbox is still coming up',
    cloud_platform: 'aws',
    aws_account: { region: 'us-east-1' },
    labels: { tier: 'trial' },
    updated_at: '2026-09-22T11:40:00Z',
    created_at: '2026-09-22T11:30:00Z',
  },
] as unknown as TOperatorInstall[]

export const operatorApprovals = [
  {
    id: 'aprvl-1',
    type: 'terraform_plan',
    workflow_step: {
      name: 'apply changes',
      owner_type: 'installs',
      owner_id: 'instb2c3d4e5f6g7h8i9j0k1l2',
      install_workflow_id: 'wkfl-approve-1',
    },
  },
  {
    id: 'aprvl-2',
    type: 'helm_approval',
    workflow_step: {
      name: 'deploy components',
      owner_type: 'installs',
      owner_id: 'instg7h8i9j0k1l2m3n4o5p6q7',
      install_workflow_id: 'wkfl-approve-2',
    },
  },
] as any

export const operatorActiveWorkflows = [
  {
    id: 'wkfl-approve-1',
    owner_type: 'installs',
    owner_id: 'instb2c3d4e5f6g7h8i9j0k1l2',
    metadata: { owner_name: 'globex-production' },
  },
  {
    id: 'wkfl-approve-2',
    owner_type: 'installs',
    owner_id: 'instg7h8i9j0k1l2m3n4o5p6q7',
    metadata: { owner_name: 'vehement-production' },
  },
] as any

export const OPERATOR_ORG_ID = ORG_ID
