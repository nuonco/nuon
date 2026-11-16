import type { TAppBranchConfig, TInstall, TInstallStatus, TInstallStatusAxis } from '@/types'

const main = { id: 'br-main', name: 'main', app_id: 'app-1' }
const monthly = { id: 'br-monthly', name: 'monthly', app_id: 'app-1' }

const install = (
  id: string,
  name: string,
  overrides: Partial<TInstall> = {}
): TInstall => ({
  id,
  name,
  org_id: 'org-1',
  app_id: 'app-1',
  labels: {},
  runner_status: 'active',
  runner_type: 'aws',
  cloud_platform: 'aws',
  sandbox_status: 'active',
  composite_component_status: 'active',
  composite_health_status: 'healthy',
  composite_health_status_description: 'All components are healthy',
  lifecycle_phase: { phase: 'provisioned' },
  created_at: '2026-09-01T12:00:00Z',
  updated_at: '2026-11-15T09:30:00Z',
  ...overrides,
})

export const mixedFleet: TInstall[] = [
  install('inst-canary', 'acme-canary', {
    labels: { env: 'stage', tier: 'canary' },
    app_branch: main,
    app_branch_group: 'stage',
  }),
  install('inst-stage-eu', 'acme-stage-eu', {
    labels: { env: 'stage', region: 'eu-west-1' },
    app_branch: main,
    app_branch_group: 'stage',
    composite_health_status: 'progressing',
    composite_health_status_description: 'A component is rolling out',
    composite_component_status: 'pending',
  }),
  install('inst-prod-us', 'acme-prod-us', {
    labels: { env: 'prod', region: 'us-east-1' },
    app_branch: main,
    app_branch_group: 'production',
  }),
  install('inst-prod-eu', 'acme-prod-eu', {
    labels: { env: 'prod', region: 'eu-west-1' },
    app_branch: main,
    app_branch_group: 'production',
    composite_health_status: 'degraded',
    composite_health_status_description: 'A component is degraded',
  }),
  install('inst-prod-ap', 'acme-prod-ap', {
    labels: { env: 'prod', region: 'ap-south-1' },
    app_branch: main,
    app_branch_group: 'production',
    runner_status: 'offline',
    runner_status_description: 'Runner has not sent a heartbeat',
    composite_health_status: 'unhealthy',
    composite_health_status_description: 'A component is unhealthy',
    composite_component_status: 'error',
  }),
  install('inst-bank', 'acme-bank', {
    labels: { env: 'prod', cadence: 'monthly' },
    app_branch: monthly,
    app_branch_group: 'production',
    cloud_platform: 'azure',
    runner_type: 'azure',
  }),
  install('inst-labs', 'acme-labs', {
    labels: { team: 'labs' },
    composite_health_status: 'unset',
    composite_health_status_description: undefined,
    sandbox_status: 'provisioning',
    runner_status: 'provisioning',
    composite_component_status: 'pending',
    lifecycle_phase: { phase: 'provisioning' },
  }),
]

export const mixedFleetConfigs: TAppBranchConfig[] = [
  {
    app_branch_id: 'br-main',
    install_groups: [
      {
        name: 'stage',
        order: 1,
        label_selector: { match_labels: { env: 'stage' } },
      },
      {
        name: 'production',
        order: 2,
        default: true,
        label_selector: { match_labels: { env: 'prod' } },
      },
    ],
  },
  {
    app_branch_id: 'br-monthly',
    install_groups: [
      {
        name: 'production',
        order: 1,
        default: true,
        label_selector: { match_labels: { cadence: 'monthly', env: 'prod' } },
      },
    ],
  },
]

const axis = (status: string, description: string): TInstallStatusAxis => ({
  status,
  status_human_description: description,
})

const deployed = axis('active', 'All deployed')
const healthy = axis('healthy', 'All healthy')

const allGood: TInstallStatus = {
  deployments: deployed,
  resources: healthy,
  health_checks: healthy,
}

export const mixedFleetStatus: Record<string, TInstallStatus> = {
  'inst-canary': allGood,
  'inst-stage-eu': {
    deployments: axis('pending', 'In progress'),
    resources: axis('progressing', 'Progressing'),
    health_checks: healthy,
  },
  'inst-prod-us': allGood,
  'inst-prod-eu': {
    deployments: deployed,
    resources: axis('degraded', 'Degraded'),
    health_checks: healthy,
  },
  'inst-prod-ap': {
    deployments: axis('error', '1 deployment failed'),
    resources: axis('unhealthy', 'Unhealthy'),
    health_checks: axis('unhealthy', 'Unhealthy'),
  },
  'inst-bank': allGood,
  'inst-labs': {
    deployments: axis('pending', 'Not deployed'),
    resources: axis('unknown', 'No resources'),
    health_checks: axis('unknown', 'No health checks'),
  },
}
