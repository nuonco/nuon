import { installAppBranchGroup } from '@/components/branches/active-app-branch-connection'
import type { TAppBranchConfig, TInstall } from '@/types'
import {
  bearsHealthVerdict,
  compareHealthSeverityDesc,
  HEALTH_DEGRADED,
  HEALTH_HEALTHY,
  HEALTH_PROGRESSING,
  HEALTH_UNHEALTHY,
  HEALTH_UNKNOWN,
  worstHealth,
} from '@/utils/health-utils'

export const NO_SIGNAL = 'no-signal'

export const HEALTH_FILTER_OPTIONS = [
  HEALTH_UNHEALTHY,
  HEALTH_DEGRADED,
  HEALTH_PROGRESSING,
  HEALTH_HEALTHY,
  HEALTH_UNKNOWN,
  NO_SIGNAL,
]

export type TFleetRunnerFilter = 'active' | 'not-active'

export type TFleetFilters = {
  search: string
  labels: Set<string>
  health?: Set<string>
  runner?: TFleetRunnerFilter
}

export type TFleetGroup = {
  key: string
  name?: string
  labels?: Record<string, string>
  installs: TInstall[]
  worst: string
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
  healthy: number
  degraded: number
  unhealthy: number
  unknown: number
  noSignal: number
  runnerNotActive: number
}

export const healthFilterValue = (install: TInstall): string => {
  const health = install.composite_health_status
  if (!health || health === 'unset' || health === 'not-applicable')
    return NO_SIGNAL
  return bearsHealthVerdict(health) ? health : HEALTH_UNKNOWN
}

// Progressing folds into unknown to match the buckets GET /v1/installs/health
// returns, so these counts agree with a rollout script polling that endpoint.
export const healthBucket = (install: TInstall): string => {
  const value = healthFilterValue(install)
  return value === HEALTH_PROGRESSING ? HEALTH_UNKNOWN : value
}

export const isRunnerActive = (install: TInstall) =>
  install.runner_status === 'active'

export const labelPairs = (install: TInstall) =>
  Object.entries(install.labels ?? {})
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([key, value]) => `${key}:${value}`)

export const labelOptions = (installs: TInstall[]) =>
  [...new Set(installs.flatMap(labelPairs))].sort()

export const matchesFilters = (install: TInstall, filters: TFleetFilters) => {
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
  return true
}

export const summarizeFleet = (installs: TInstall[]): TFleetSummary => {
  const summary: TFleetSummary = {
    total: installs.length,
    healthy: 0,
    degraded: 0,
    unhealthy: 0,
    unknown: 0,
    noSignal: 0,
    runnerNotActive: 0,
  }
  for (const install of installs) {
    const bucket = healthBucket(install)
    if (bucket === HEALTH_HEALTHY) summary.healthy++
    else if (bucket === HEALTH_DEGRADED) summary.degraded++
    else if (bucket === HEALTH_UNHEALTHY) summary.unhealthy++
    else if (bucket === NO_SIGNAL) summary.noSignal++
    else summary.unknown++
    if (!isRunnerActive(install)) summary.runnerNotActive++
  }
  return summary
}

const compareInstalls = (a: TInstall, b: TInstall) =>
  compareHealthSeverityDesc(
    a.composite_health_status,
    b.composite_health_status
  ) || (a.name ?? '').localeCompare(b.name ?? '')

const groupsOf = (
  installs: TInstall[],
  config?: TAppBranchConfig
): TFleetGroup[] => {
  const byGroup = new Map<string, TInstall[]>()
  for (const install of installs) {
    const key = installAppBranchGroup(install) ?? ''
    byGroup.set(key, [...(byGroup.get(key) ?? []), install])
  }
  const order = (name: string) =>
    config?.install_groups?.find((group) => group.name === name)?.order ??
    Number.MAX_SAFE_INTEGER
  return [...byGroup.entries()]
    .sort(([a], [b]) =>
      !a ? 1 : !b ? -1 : order(a) - order(b) || a.localeCompare(b)
    )
    .map(([key, members]) => ({
      key: key || 'no-group',
      name: key || undefined,
      labels: config?.install_groups?.find((group) => group.name === key)
        ?.label_selector?.match_labels,
      installs: [...members].sort(compareInstalls),
      worst: worstHealth(members.map((m) => m.composite_health_status)),
    }))
}

export const groupFleet = (
  installs: TInstall[],
  configs: TAppBranchConfig[] = []
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
        configs.find((config) => config.app_branch_id === key)
      ),
    }))
    .sort((a, b) =>
      !a.branchId ? 1 : !b.branchId ? -1 : (a.name ?? '').localeCompare(b.name ?? '')
    )
}
