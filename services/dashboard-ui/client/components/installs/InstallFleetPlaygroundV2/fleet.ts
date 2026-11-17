import { installAppBranchGroup } from '@/components/branches/active-app-branch-connection'
import type {
  TAppBranch,
  TAppBranchConfig,
  TAppBranchRun,
  TInstall,
  TInstallOverview,
} from '@/types'
import {
  bearsHealthVerdict,
  compareHealthSeverityDesc,
  HEALTH_DEGRADED,
  HEALTH_HEALTHY,
  HEALTH_PROGRESSING,
  HEALTH_UNHEALTHY,
  HEALTH_UNKNOWN,
} from '@/utils/health-utils'
import { humanize } from '@/utils/string-utils'

export const NO_SIGNAL = 'no-signal'

export const HEALTH_FILTER_OPTIONS = [
  HEALTH_UNHEALTHY,
  HEALTH_DEGRADED,
  HEALTH_PROGRESSING,
  HEALTH_HEALTHY,
  HEALTH_UNKNOWN,
  NO_SIGNAL,
]

export const LATEST_RUN = 'latest'

export type TFleetRunnerFilter = 'active' | 'not-active'
export type TRolloutState = 'applied' | 'behind' | 'unknown'
export type TRunMatch = 'on' | 'off'

export type TFleetRunRef = { branchId: string; appConfigId?: string }

export type TFleetContext = {
  latestConfigIdByBranch?: Record<string, string | undefined>
  runsById?: Record<string, TFleetRunRef>
  overviewById?: Record<string, TInstallOverview>
  runByBranch?: Record<string, string | undefined>
  scope?: string
}

export type TFleetFilters = {
  search: string
  labels: Set<string>
  health?: Set<string>
  runner?: TFleetRunnerFilter
  rolloutByBranch?: Record<
    string,
    Exclude<TRolloutState, 'unknown'> | undefined
  >
}

export type TFleetGroup = {
  key: string
  name?: string
  labels?: Record<string, string>
  installs: TInstall[]
}

export type TFleetBranch = {
  key: string
  branchId?: string
  name?: string
  installs: TInstall[]
  groups: TFleetGroup[]
}

export type TFleetSummary = {
  total: number
  applied: number
  behind: number
  runnerNotActive: number
}

export const healthFilterValue = (install: TInstall): string => {
  const health = install.composite_health_status
  if (!health || health === 'unset' || health === 'not-applicable')
    return NO_SIGNAL
  return bearsHealthVerdict(health) ? health : HEALTH_UNKNOWN
}

export const isRunnerActive = (install: TInstall) =>
  install.runner_status === 'active'

export const latestConfigIdByBranch = (branches: TAppBranch[] = []) =>
  Object.fromEntries(
    branches
      .filter((branch) => branch.id)
      .map((branch) => [branch.id as string, branch.latest_run?.app_config_id])
  )

export const rolloutState = (
  install: TInstall,
  context: TFleetContext = {}
): TRolloutState => {
  const run = context.runByBranch?.[install.app_branch?.id ?? ''] ?? LATEST_RUN
  const match = runMatch(install, run, context.scope, context)
  return match === 'on' ? 'applied' : match === 'off' ? 'behind' : 'unknown'
}

const ROLLOUT_RANK: Record<TRolloutState, number> = {
  behind: 0,
  unknown: 1,
  applied: 2,
}

export const runsIndex = (
  branchRuns: Record<string, TAppBranchRun[]> = {}
): Record<string, TFleetRunRef> =>
  Object.fromEntries(
    Object.entries(branchRuns).flatMap(([branchId, runs]) =>
      runs
        .filter((run) => run.id)
        .map((run) => [
          run.id as string,
          { branchId, appConfigId: run.app_config_id },
        ])
    )
  )

const runTargetConfig = (
  install: TInstall,
  run: string,
  context: TFleetContext
) => {
  const branchId = install.app_branch?.id
  if (!branchId) return undefined
  if (run === LATEST_RUN) return context.latestConfigIdByBranch?.[branchId]
  const ref = context.runsById?.[run]
  return ref?.branchId === branchId ? ref.appConfigId : undefined
}

const appliedConfig = (
  install: TInstall,
  scope: string | undefined,
  context: TFleetContext
) =>
  scope
    ? context.overviewById?.[install.id ?? '']?.config_drift?.components?.find(
        (component) => component.component_id === scope
      )?.applied_app_config_id
    : install.app_config_id

export const runMatch = (
  install: TInstall,
  run: string,
  scope: string | undefined,
  context: TFleetContext = {}
): TRunMatch | undefined => {
  const target = runTargetConfig(install, run, context)
  const applied = appliedConfig(install, scope, context)
  if (!target || !applied) return undefined
  return applied === target ? 'on' : 'off'
}

export const countOnRun = (
  installs: TInstall[],
  run: string,
  scope: string | undefined,
  context: TFleetContext = {}
) =>
  installs.filter((install) => runMatch(install, run, scope, context) === 'on')
    .length

export const componentOptions = (
  overviewById: Record<string, TInstallOverview> = {}
) => {
  const names = new Map<string, string>()
  for (const overview of Object.values(overviewById)) {
    for (const component of overview.config_drift?.components ?? []) {
      if (component.component_id)
        names.set(
          component.component_id,
          component.name ?? component.component_id
        )
    }
  }
  return [...names.entries()]
    .map(([value, label]) => ({ value, label }))
    .sort((a, b) => a.label.localeCompare(b.label))
}

export const labelPairs = (install: TInstall) =>
  Object.entries(install.labels ?? {})
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => `${key}:${value}`)

export const labelOptions = (installs: TInstall[]) =>
  [...new Set(installs.flatMap(labelPairs))].sort()

export const healthFilterLabel = (value: string) =>
  value === NO_SIGNAL ? 'No health signal' : humanize(value)

export const matchesFilters = (
  install: TInstall,
  filters: TFleetFilters,
  context: TFleetContext = {}
) => {
  const needle = filters.search.trim().toLowerCase()
  if (
    needle &&
    !install.name?.toLowerCase().includes(needle) &&
    !install.id?.toLowerCase().includes(needle)
  )
    return false
  if (filters.labels.size) {
    const have = new Set(labelPairs(install))
    if (![...filters.labels].every((label) => have.has(label))) return false
  }
  if (filters.health?.size && !filters.health.has(healthFilterValue(install)))
    return false
  if (filters.runner === 'active' && !isRunnerActive(install)) return false
  if (filters.runner === 'not-active' && isRunnerActive(install)) return false
  const rollout = filters.rolloutByBranch?.[install.app_branch?.id ?? '']
  if (rollout && rolloutState(install, context) !== rollout) return false
  return true
}

export const summarizeFleet = (
  installs: TInstall[],
  context: TFleetContext = {}
): TFleetSummary => {
  const summary: TFleetSummary = {
    total: installs.length,
    applied: 0,
    behind: 0,
    runnerNotActive: 0,
  }
  const latest = { ...context, runByBranch: undefined, scope: undefined }
  for (const install of installs) {
    const state = rolloutState(install, latest)
    if (state === 'applied') summary.applied++
    else if (state === 'behind') summary.behind++
    if (!isRunnerActive(install)) summary.runnerNotActive++
  }
  return summary
}

const compareInstalls =
  (context: TFleetContext) => (a: TInstall, b: TInstall) =>
    ROLLOUT_RANK[rolloutState(a, context)] -
      ROLLOUT_RANK[rolloutState(b, context)] ||
    compareHealthSeverityDesc(
      a.composite_health_status,
      b.composite_health_status
    ) ||
    (a.name ?? '').localeCompare(b.name ?? '')

const groupsOf = (
  installs: TInstall[],
  config: TAppBranchConfig | undefined,
  context: TFleetContext
): TFleetGroup[] => {
  const byGroup = new Map<string, TInstall[]>()
  for (const install of installs) {
    const key = installAppBranchGroup(install) ?? ''
    byGroup.set(key, [...(byGroup.get(key) ?? []), install])
  }
  const order = (name: string) =>
    config?.install_groups?.find((group) => group.name === name)?.order ??
    Number.MAX_SAFE_INTEGER
  const compare = compareInstalls(context)
  return [...byGroup.entries()]
    .sort(([a], [b]) =>
      !a ? 1 : !b ? -1 : order(a) - order(b) || a.localeCompare(b)
    )
    .map(([key, members]) => ({
      key: key || 'no-group',
      name: key || undefined,
      labels: config?.install_groups?.find((group) => group.name === key)
        ?.label_selector?.match_labels,
      installs: [...members].sort(compare),
    }))
}

export const groupFleet = (
  installs: TInstall[],
  configs: TAppBranchConfig[] = [],
  context: TFleetContext = {}
): TFleetBranch[] => {
  const byBranch = new Map<string, TInstall[]>()
  for (const install of installs) {
    const key = install.app_branch?.id ?? ''
    byBranch.set(key, [...(byBranch.get(key) ?? []), install])
  }
  return [...byBranch.entries()]
    .map(([key, members]) => ({
      key: key || 'no-branch',
      branchId: key || undefined,
      name: members[0]?.app_branch?.name,
      installs: members,
      groups: groupsOf(
        members,
        configs.find((config) => config.app_branch_id === key),
        context
      ),
    }))
    .sort((a, b) =>
      !a.branchId
        ? 1
        : !b.branchId
          ? -1
          : (a.name ?? '').localeCompare(b.name ?? '')
    )
}
