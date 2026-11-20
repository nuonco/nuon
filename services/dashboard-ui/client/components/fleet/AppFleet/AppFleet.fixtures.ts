import type {
  TAppFleet,
  TAppFleetCounts,
  TAppFleetGroup,
  TAppFleetInstall,
} from '@/types'

const HEALTH = [
  'healthy',
  'healthy',
  'degraded',
  'unhealthy',
  'progressing',
  'no-signal',
]
const RUNNER = ['active', 'active', 'active', 'offline']
const ROLLOUT = ['applied', 'applied', 'behind', 'unknown']

const install = (
  branchId: string,
  group: string | undefined,
  index: number
): TAppFleetInstall => ({
  id: `inst${branchId}${group ?? 'none'}${index}`.padEnd(26, '0').slice(0, 26),
  name: `acme-${group ?? 'adhoc'}-${index + 1}`,
  app_branch_id: branchId,
  app_branch_group: group,
  rollout: ROLLOUT[index % ROLLOUT.length],
  health: HEALTH[index % HEALTH.length],
  runner_status: RUNNER[index % RUNNER.length],
  app_config_id: `cfg-${index}`,
  applied_config_id: `cfg-${index}`,
  labels: { env: group === 'canary' ? 'stage' : 'prod', region: 'us-east-1' },
})

const countsOf = (installs: TAppFleetInstall[], total = installs.length) => {
  const counts: TAppFleetCounts = {
    total,
    matched: total,
    applied: 0,
    behind: 0,
    unknown: 0,
  }
  const scale = total / Math.max(installs.length, 1)
  for (const item of installs) {
    const key = (item.rollout ?? 'unknown') as 'applied' | 'behind' | 'unknown'
    counts[key] = (counts[key] ?? 0) + scale
  }
  return {
    ...counts,
    applied: Math.round(counts.applied ?? 0),
    behind: Math.round(counts.behind ?? 0),
    unknown:
      total - Math.round(counts.applied ?? 0) - Math.round(counts.behind ?? 0),
  }
}

const group = (
  branchId: string,
  name: string | undefined,
  size: number,
  shown = Math.min(size, 10),
  order?: number,
  match_labels?: Record<string, string>
): TAppFleetGroup => {
  const installs = Array.from({ length: shown }, (_, i) =>
    install(branchId, name, i)
  )
  return {
    name,
    order,
    match_labels,
    counts: countsOf(installs, size),
    installs,
    has_more: size > shown,
  }
}

const sumCounts = (groups: TAppFleetGroup[]): TAppFleetCounts =>
  groups.reduce<TAppFleetCounts>(
    (acc, g) => ({
      total: (acc.total ?? 0) + (g.counts?.total ?? 0),
      matched: (acc.matched ?? 0) + (g.counts?.matched ?? 0),
      applied: (acc.applied ?? 0) + (g.counts?.applied ?? 0),
      behind: (acc.behind ?? 0) + (g.counts?.behind ?? 0),
      unknown: (acc.unknown ?? 0) + (g.counts?.unknown ?? 0),
    }),
    {}
  )

const mainGroups = [
  group('brmain', 'canary', 3, 3, 0, { tier: 'canary' }),
  group('brmain', 'production', 42, 10, 1, { env: 'prod' }),
  group('brmain', undefined, 2),
]
const releaseGroups = [
  group('brrelease', 'early-access', 6, 6, 0, { tier: 'ea' }),
]
const noBranchGroups = [group('', undefined, 1)]

const allGroups = [...mainGroups, ...releaseGroups, ...noBranchGroups]
const total = sumCounts(allGroups)

export const appFleet: TAppFleet = {
  summary: { ...total, runner_not_active: 13 },
  branches: [
    {
      id: 'brmain',
      name: 'main',
      target_run: { id: 'run-main', app_config_id: 'cfg-0' },
      counts: sumCounts(mainGroups),
      groups: mainGroups,
    },
    {
      id: 'brrelease',
      name: 'release/2.x',
      target_run: { id: 'run-release', app_config_id: 'cfg-0' },
      counts: sumCounts(releaseGroups),
      groups: releaseGroups,
    },
    { counts: sumCounts(noBranchGroups), groups: noBranchGroups },
  ],
}

const filteredMain = mainGroups.map((g) => ({
  ...g,
  counts: { ...g.counts, matched: g.name === 'production' ? 4 : 0 },
  installs: g.name === 'production' ? (g.installs ?? []).slice(0, 4) : [],
  has_more: false,
}))

export const filteredAppFleet: TAppFleet = {
  summary: { ...total, matched: 4, runner_not_active: 1 },
  branches: [
    {
      ...appFleet.branches![0],
      counts: { ...sumCounts(mainGroups), matched: 4 },
      groups: filteredMain,
    },
    {
      ...appFleet.branches![1],
      counts: { ...sumCounts(releaseGroups), matched: 0 },
    },
    {
      ...appFleet.branches![2],
      counts: { ...sumCounts(noBranchGroups), matched: 0 },
    },
  ],
}

export const emptyAppFleet: TAppFleet = {
  summary: {
    total: 0,
    matched: 0,
    applied: 0,
    behind: 0,
    unknown: 0,
    runner_not_active: 0,
  },
  branches: [],
}
