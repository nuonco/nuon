import type {
  TAppBranch,
  TAppBranchConfig,
  TAppBranchRun,
  TInstall,
  TInstallOverview,
  TInstallStatus,
  TInstallStatusAxis,
  TWorkflowStepApproval,
} from '@/types'

const main = { id: 'br-main', name: 'main', app_id: 'app-1' }
const monthly = { id: 'br-monthly', name: 'monthly', app_id: 'app-1' }

const MAIN_LATEST = 'cfg-main-3'
const MAIN_STALE = 'cfg-main-2'
const MONTHLY_LATEST = 'cfg-monthly-2'

const COMPONENTS = [
  { id: 'comp-api', name: 'api' },
  { id: 'comp-web', name: 'web' },
  { id: 'comp-worker', name: 'worker' },
]

const overviewFor = (
  current: string,
  stale: string,
  drifted: string[] = []
): TInstallOverview => ({
  config_drift: {
    current_app_config_id: current,
    components: COMPONENTS.map((component) => ({
      component_id: component.id,
      name: component.name,
      applied_app_config_id: drifted.includes(component.id) ? stale : current,
      drifted: drifted.includes(component.id),
    })),
  },
})

const branchRun = (
  id: string,
  appConfigId: string,
  message: string,
  sha: string,
  createdAt: string
): TAppBranchRun => ({
  id,
  app_config_id: appConfigId,
  head_sha: sha,
  status: 'active',
  created_at: createdAt,
  vcs_connection_commit: { sha, message },
})

const branchWithRun = (
  branch: { id: string; name: string; app_id: string },
  latest: TAppBranchRun
): TAppBranch => ({
  ...branch,
  latest_run: latest,
})

const ROLES = [
  { type: 'runner_provision', display_name: 'Provision' },
  { type: 'runner_maintenance', display_name: 'Maintenance' },
  { type: 'runner_deprovision', display_name: 'Deprovision' },
] as const

const installRoles = (
  installId: string,
  disabled: (typeof ROLES)[number]['type'][] = []
): TInstall['install_roles'] =>
  ROLES.map((role) => ({
    id: `role-${installId}-${role.type}`,
    install_id: installId,
    role_id: `acme-${role.type}`,
    provisioned: true,
    enabled: !disabled.includes(role.type),
    app_role_config: { type: role.type, display_name: role.display_name },
  }))

const hex = (value: string) =>
  [...value]
    .map((char) => char.charCodeAt(0).toString(16).padStart(2, '0'))
    .join('')

const overrideInputs = (
  installId: string,
  components: string[]
): TInstall['install_inputs'] => [
  {
    id: `inputs-${installId}`,
    install_id: installId,
    values: Object.fromEntries(
      components.map((component) => [
        `nuon_component_override_v1_helm_values_${hex(component)}`,
        'replicaCount: 2\n',
      ])
    ),
  },
]

const RECENT_UPGRADE = '2026-11-16T15:20:00Z'
const STALE_UPGRADE = '2026-10-09T10:05:00Z'

const appliedRef = (configId: string, at: string) => ({
  applied_config_id: configId,
  applied_config_at: at,
})

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
  install_roles: installRoles(id),
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
    app_config_id: MAIN_LATEST,
    app_config_ref: appliedRef(MAIN_LATEST, RECENT_UPGRADE),
    app_branch: main,
    app_branch_group: 'stage',
    install_inputs: overrideInputs('inst-canary', ['api', 'web']),
  }),
  install('inst-stage-eu', 'acme-stage-eu', {
    labels: { env: 'stage', region: 'eu-west-1' },
    app_config_id: MAIN_LATEST,
    app_config_ref: appliedRef(MAIN_LATEST, RECENT_UPGRADE),
    app_branch: main,
    app_branch_group: 'stage',
    composite_health_status: 'progressing',
    composite_health_status_description: 'A component is rolling out',
    composite_component_status: 'pending',
  }),
  install('inst-prod-us', 'acme-prod-us', {
    labels: { env: 'prod', region: 'us-east-1' },
    app_config_id: MAIN_LATEST,
    app_config_ref: appliedRef(MAIN_LATEST, RECENT_UPGRADE),
    app_branch: main,
    app_branch_group: 'production',
    install_inputs: overrideInputs('inst-prod-us', ['api']),
  }),
  install('inst-prod-eu', 'acme-prod-eu', {
    labels: { env: 'prod', region: 'eu-west-1' },
    app_config_id: MAIN_STALE,
    app_config_ref: appliedRef(MAIN_STALE, STALE_UPGRADE),
    app_branch: main,
    app_branch_group: 'production',
    composite_health_status: 'degraded',
    composite_health_status_description: 'A component is degraded',
    install_roles: installRoles('inst-prod-eu', ['runner_maintenance']),
  }),
  install('inst-prod-ap', 'acme-prod-ap', {
    labels: { env: 'prod', region: 'ap-south-1' },
    app_config_id: MAIN_STALE,
    app_config_ref: appliedRef(MAIN_STALE, STALE_UPGRADE),
    app_branch: main,
    app_branch_group: 'production',
    runner_status: 'offline',
    runner_status_description: 'Runner has not sent a heartbeat',
    composite_health_status: 'unhealthy',
    composite_health_status_description: 'A component is unhealthy',
    composite_component_status: 'error',
    install_roles: installRoles('inst-prod-ap', [
      'runner_provision',
      'runner_deprovision',
    ]),
  }),
  install('inst-bank', 'acme-bank', {
    labels: { env: 'prod', cadence: 'monthly' },
    app_config_id: MONTHLY_LATEST,
    app_config_ref: appliedRef(MONTHLY_LATEST, RECENT_UPGRADE),
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

const mainLatestRun = branchRun(
  'run-main-3',
  MAIN_LATEST,
  'Bump api to v2.4.0',
  'a1b2c3d4e5f6',
  '2026-11-16T15:20:00Z'
)
const mainStaleRun = branchRun(
  'run-main-2',
  MAIN_STALE,
  'Bump api to v2.3.1',
  'f6e5d4c3b2a1',
  '2026-11-09T10:05:00Z'
)
const monthlyLatestRun = branchRun(
  'run-monthly-2',
  MONTHLY_LATEST,
  'November release',
  '0a9b8c7d6e5f',
  '2026-11-02T13:00:00Z'
)

export const mixedFleetBranches: TAppBranch[] = [
  branchWithRun(main, mainLatestRun),
  branchWithRun(monthly, monthlyLatestRun),
]

export const mixedFleetRuns: Record<string, TAppBranchRun[]> = {
  'br-main': [mainLatestRun, mainStaleRun],
  'br-monthly': [monthlyLatestRun],
}

export const mixedFleetOverviews: Record<string, TInstallOverview> = {
  'inst-canary': overviewFor(MAIN_LATEST, MAIN_STALE),
  'inst-stage-eu': overviewFor(MAIN_LATEST, MAIN_STALE),
  'inst-prod-us': overviewFor(MAIN_LATEST, MAIN_STALE, ['comp-api']),
  'inst-prod-eu': overviewFor(MAIN_STALE, MAIN_STALE),
  'inst-prod-ap': overviewFor(MAIN_STALE, MAIN_STALE),
  'inst-bank': overviewFor(MONTHLY_LATEST, 'cfg-monthly-1'),
}

type TLargeGroup = {
  name: string
  matchLabels: Record<string, string>
  count: number
  progress: number
}

type TLargeBranch = {
  id: string
  name: string
  latestConfigId: string
  staleConfigId: string
  latestMessage: string
  staleMessage: string
  groups: TLargeGroup[]
}

const LARGE_BRANCHES: TLargeBranch[] = [
  {
    id: 'br-main',
    name: 'main',
    latestConfigId: 'cfg-main-9',
    staleConfigId: 'cfg-main-8',
    latestMessage: 'Bump api to v2.4.0',
    staleMessage: 'Bump api to v2.3.1',
    groups: [
      {
        name: 'canary',
        matchLabels: { ring: 'canary' },
        count: 4,
        progress: 1,
      },
      { name: 'stage', matchLabels: { env: 'stage' }, count: 8, progress: 1 },
      {
        name: 'prod-us',
        matchLabels: { env: 'prod', geo: 'us' },
        count: 28,
        progress: 0.7,
      },
      {
        name: 'prod-eu',
        matchLabels: { env: 'prod', geo: 'eu' },
        count: 22,
        progress: 0.4,
      },
      {
        name: 'prod-ap',
        matchLabels: { env: 'prod', geo: 'ap' },
        count: 12,
        progress: 0,
      },
    ],
  },
  {
    id: 'br-monthly',
    name: 'monthly',
    latestConfigId: 'cfg-monthly-4',
    staleConfigId: 'cfg-monthly-3',
    latestMessage: 'November release',
    staleMessage: 'October release',
    groups: [
      {
        name: 'wave-1',
        matchLabels: { cadence: 'monthly', wave: '1' },
        count: 6,
        progress: 1,
      },
      {
        name: 'wave-2',
        matchLabels: { cadence: 'monthly', wave: '2' },
        count: 14,
        progress: 0.5,
      },
      {
        name: 'wave-3',
        matchLabels: { cadence: 'monthly', wave: '3' },
        count: 18,
        progress: 0,
      },
    ],
  },
  {
    id: 'br-lts',
    name: 'release/lts',
    latestConfigId: 'cfg-lts-2',
    staleConfigId: 'cfg-lts-1',
    latestMessage: 'Patch worker base image',
    staleMessage: 'LTS 1.0 baseline',
    groups: [
      {
        name: 'regulated',
        matchLabels: { tier: 'regulated' },
        count: 9,
        progress: 0.3,
      },
      {
        name: 'production',
        matchLabels: { tier: 'lts' },
        count: 11,
        progress: 1,
      },
    ],
  },
  {
    id: 'br-preview',
    name: 'preview',
    latestConfigId: 'cfg-preview-7',
    staleConfigId: 'cfg-preview-6',
    latestMessage: 'Try new web layout',
    staleMessage: 'Add feature flags',
    groups: [
      {
        name: 'internal',
        matchLabels: { ring: 'internal' },
        count: 3,
        progress: 1,
      },
      {
        name: 'design-partners',
        matchLabels: { ring: 'partner' },
        count: 5,
        progress: 0.6,
      },
    ],
  },
]

const LARGE_UNASSIGNED = 7

const REGIONS = [
  'us-east-1',
  'us-west-2',
  'eu-west-1',
  'eu-central-1',
  'ap-south-1',
]

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
  const overviews: Record<string, TInstallOverview> = {}
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
        install_roles: installRoles(
          id,
          n % 13 === 0
            ? ['runner_maintenance']
            : n % 19 === 0
              ? ['runner_deprovision']
              : []
        ),
        install_inputs: n % 7 === 0 ? overrideInputs(id, ['api']) : undefined,
        updated_at: `2026-11-${String(1 + (n % 15)).padStart(2, '0')}T09:30:00Z`,
        ...overrides,
      })
    )
    status[id] = statusOverride ?? installStatus
  }

  for (const branch of LARGE_BRANCHES) {
    for (const group of branch.groups) {
      const appliedCount = Math.round(group.count * group.progress)
      for (let i = 0; i < group.count; i++) {
        const isApplied = i < appliedCount
        push(group.matchLabels, {
          app_branch: { id: branch.id, name: branch.name, app_id: 'app-1' },
          app_branch_group: group.name,
          app_config_id: isApplied
            ? branch.latestConfigId
            : branch.staleConfigId,
          app_config_ref: appliedRef(
            isApplied ? branch.latestConfigId : branch.staleConfigId,
            isApplied ? RECENT_UPGRADE : STALE_UPGRADE
          ),
        })
        overviews[`inst-${String(n).padStart(3, '0')}`] = isApplied
          ? overviewFor(
              branch.latestConfigId,
              branch.staleConfigId,
              n % 3 === 0 ? ['comp-api'] : []
            )
          : overviewFor(branch.staleConfigId, branch.staleConfigId)
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

  const runs: Record<string, TAppBranchRun[]> = Object.fromEntries(
    LARGE_BRANCHES.map((branch, i) => [
      branch.id,
      [
        branchRun(
          `run-${branch.id}-latest`,
          branch.latestConfigId,
          branch.latestMessage,
          `${'abcdef'[i]}${i}c0ffee${i}1234`,
          `2026-11-${String(16 - i).padStart(2, '0')}T15:20:00Z`
        ),
        branchRun(
          `run-${branch.id}-prev`,
          branch.staleConfigId,
          branch.staleMessage,
          `${'abcdef'[i]}${i}deadbe${i}5678`,
          `2026-11-${String(9 - i).padStart(2, '0')}T10:05:00Z`
        ),
        ...(branch.id === 'br-main'
          ? [
              'Bump api to v2.3.0',
              'Rotate worker credentials',
              'Bump web to v1.8.2',
              'Add stage sandbox',
              'Bump api to v2.2.4',
            ].map((message, n) =>
              branchRun(
                `run-br-main-old-${n}`,
                `cfg-main-${7 - n}`,
                message,
                `a${n}0ld${n}feed${n}9abc`,
                `2026-10-${String(28 - n * 4).padStart(2, '0')}T09:00:00Z`
              )
            )
          : []),
      ],
    ])
  )

  const branches: TAppBranch[] = LARGE_BRANCHES.map((branch) =>
    branchWithRun(
      { id: branch.id, name: branch.name, app_id: 'app-1' },
      runs[branch.id][0]
    )
  )

  return { installs, configs, status, branches, overviews, runs }
}

const largeFleetData = buildLargeFleet()

export const largeFleet = largeFleetData.installs
export const largeFleetConfigs = largeFleetData.configs
export const largeFleetStatus = largeFleetData.status
export const largeFleetBranches = largeFleetData.branches
export const largeFleetOverviews = largeFleetData.overviews
export const largeFleetRuns = largeFleetData.runs

const versionsOf = (configIds: (string | undefined)[]) =>
  Object.fromEntries(
    configIds
      .filter((id): id is string => !!id)
      .map((id) => [id, Number(id.split('-').pop())])
  )

export const mixedFleetConfigVersions = versionsOf(
  mixedFleet.map((install) => install.app_config_id)
)

export const largeFleetConfigVersions = versionsOf(
  largeFleet.map((install) => install.app_config_id)
)

const approval = (
  id: string,
  installId: string,
  stepName: string,
  type: TWorkflowStepApproval['type']
): TWorkflowStepApproval => ({
  id,
  owner_type: 'installs',
  owner_id: installId,
  type,
  created_at: '2026-11-17T13:40:00Z',
  workflow_step: {
    id: `step-${id}`,
    name: stepName,
    owner_type: 'installs',
    owner_id: installId,
    install_workflow_id: `wf-${id}`,
  },
})

export const mixedFleetApprovals: TWorkflowStepApproval[] = [
  approval('apr-canary', 'inst-canary', 'deploy api', 'terraform_plan'),
  approval('apr-prod-eu', 'inst-prod-eu', 'deploy web', 'helm_approval'),
]

export const largeFleetApprovals: TWorkflowStepApproval[] = [5, 18, 27, 44, 61]
  .map((n) => `inst-${String(n).padStart(3, '0')}`)
  .map((installId, i) =>
    approval(
      `apr-${installId}`,
      installId,
      i % 2 ? 'deploy worker' : 'deploy api',
      i % 2 ? 'kubernetes_manifest_approval' : 'terraform_plan'
    )
  )
