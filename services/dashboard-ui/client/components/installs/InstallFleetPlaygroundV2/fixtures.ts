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

type TLargeGroup = {
  name: string
  matchLabels: Record<string, string>
  count: number
}

type TLargeBranch = { id: string; name: string; groups: TLargeGroup[] }

const LARGE_BRANCHES: TLargeBranch[] = [
  {
    id: 'br-main',
    name: 'main',
    groups: [
      { name: 'canary', matchLabels: { ring: 'canary' }, count: 4 },
      { name: 'stage', matchLabels: { env: 'stage' }, count: 8 },
      { name: 'prod-us', matchLabels: { env: 'prod', geo: 'us' }, count: 28 },
      { name: 'prod-eu', matchLabels: { env: 'prod', geo: 'eu' }, count: 22 },
      { name: 'prod-ap', matchLabels: { env: 'prod', geo: 'ap' }, count: 12 },
    ],
  },
  {
    id: 'br-monthly',
    name: 'monthly',
    groups: [
      { name: 'wave-1', matchLabels: { cadence: 'monthly', wave: '1' }, count: 6 },
      { name: 'wave-2', matchLabels: { cadence: 'monthly', wave: '2' }, count: 14 },
      { name: 'wave-3', matchLabels: { cadence: 'monthly', wave: '3' }, count: 18 },
    ],
  },
  {
    id: 'br-lts',
    name: 'release/lts',
    groups: [
      { name: 'regulated', matchLabels: { tier: 'regulated' }, count: 9 },
      { name: 'production', matchLabels: { tier: 'lts' }, count: 11 },
    ],
  },
  {
    id: 'br-preview',
    name: 'preview',
    groups: [
      { name: 'internal', matchLabels: { ring: 'internal' }, count: 3 },
      { name: 'design-partners', matchLabels: { ring: 'partner' }, count: 5 },
    ],
  },
]

const LARGE_UNASSIGNED = 7

const REGIONS = ['us-east-1', 'us-west-2', 'eu-west-1', 'eu-central-1', 'ap-south-1']

type TLargeHealth = {
  health: string
  runner: string
  status: TInstallStatus
}

const largeHealth = (n: number): TLargeHealth => {
  if (n % 23 === 0)
    return {
      health: 'unhealthy',
      runner: n % 2 ? 'offline' : 'active',
      status: {
        deployments: axis('error', '1 deployment failed'),
        resources: axis('unhealthy', 'Unhealthy'),
        health_checks: axis('unhealthy', '2 checks failing'),
      },
    }
  if (n % 11 === 0)
    return {
      health: 'degraded',
      runner: 'active',
      status: {
        deployments: deployed,
        resources: axis('degraded', 'Degraded'),
        health_checks: healthy,
      },
    }
  if (n % 17 === 0)
    return {
      health: 'progressing',
      runner: 'active',
      status: {
        deployments: axis('pending', 'In progress'),
        resources: axis('progressing', 'Progressing'),
        health_checks: healthy,
      },
    }
  return { health: 'healthy', runner: 'active', status: allGood }
}

const buildLargeFleet = () => {
  const installs: TInstall[] = []
  const status: Record<string, TInstallStatus> = {}
  let n = 0
  const push = (
    labels: Record<string, string>,
    overrides: Partial<TInstall> = {},
    statusOverride?: TInstallStatus
  ) => {
    n++
    const id = `inst-${String(n).padStart(3, '0')}`
    const region = REGIONS[n % REGIONS.length]
    const { health, runner, status: installStatus } = largeHealth(n)
    installs.push(
      install(id, `acme-${String(n).padStart(3, '0')}-${region}`, {
        labels: { ...labels, region },
        composite_health_status: health,
        runner_status: runner,
        updated_at: `2026-11-${String(1 + (n % 15)).padStart(2, '0')}T09:30:00Z`,
        ...overrides,
      })
    )
    status[id] = statusOverride ?? installStatus
  }

  for (const branch of LARGE_BRANCHES) {
    for (const group of branch.groups) {
      for (let i = 0; i < group.count; i++) {
        push(group.matchLabels, {
          app_branch: { id: branch.id, name: branch.name, app_id: 'app-1' },
          app_branch_group: group.name,
        })
      }
    }
  }
  for (let i = 0; i < LARGE_UNASSIGNED; i++) {
    const noSignal = i % 3 !== 0
    push(
      { team: 'labs' },
      {
        composite_health_status: noSignal ? 'unset' : 'healthy',
        runner_status: i % 4 ? 'active' : 'provisioning',
      },
      noSignal
        ? {
            deployments: axis('pending', 'Not deployed'),
            resources: axis('unknown', 'No resources'),
            health_checks: axis('unknown', 'No health checks'),
          }
        : allGood
    )
  }

  const configs: TAppBranchConfig[] = LARGE_BRANCHES.map((branch) => ({
    app_branch_id: branch.id,
    install_groups: branch.groups.map((group, i) => ({
      name: group.name,
      order: i + 1,
      label_selector: { match_labels: group.matchLabels },
    })),
  }))

  return { installs, configs, status }
}

const largeFleetData = buildLargeFleet()

export const largeFleet = largeFleetData.installs
export const largeFleetConfigs = largeFleetData.configs
export const largeFleetStatus = largeFleetData.status
