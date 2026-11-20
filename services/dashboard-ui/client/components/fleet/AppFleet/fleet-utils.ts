import type {
  TAppFleetCounts,
  TAppFleetFilters,
  TAppFleetRollout,
  TAppFleetRunnerFilter,
} from '@/types'
import {
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

export const healthFilterLabel = (value: string) =>
  value === NO_SIGNAL ? 'No health signal' : humanize(value)

export const ROLLOUT_OPTIONS: { value: TAppFleetRollout; label: string }[] = [
  { value: 'applied', label: 'Applied' },
  { value: 'behind', label: 'Not applied' },
  { value: 'unknown', label: 'Not tracked' },
]

export type TFleetView = 'table' | 'board'

export const FLEET_FILTER_KEYS = [
  'q',
  'labels',
  'health',
  'runner_status',
  'rollout',
  'component_id',
] as const

const isRollout = (value?: string | null): value is TAppFleetRollout =>
  value === 'applied' || value === 'behind' || value === 'unknown'

const isRunner = (value?: string | null): value is TAppFleetRunnerFilter =>
  value === 'active' || value === 'not-active'

export const parseFleetFilters = (
  params: URLSearchParams
): TAppFleetFilters => {
  const rollout = params.get('rollout')
  const runner = params.get('runner_status')
  return {
    q: params.get('q')?.trim() || undefined,
    labels: params.get('labels') || undefined,
    health: params.get('health') || undefined,
    runner_status: isRunner(runner) ? runner : undefined,
    rollout: isRollout(rollout) ? rollout : undefined,
    component_id: params.get('component_id') || undefined,
  }
}

export const parseFleetView = (params: URLSearchParams): TFleetView =>
  params.get('view') === 'board' ? 'board' : 'table'

export const hasActiveFilters = (filters: TAppFleetFilters) =>
  FLEET_FILTER_KEYS.some((key) => !!filters[key])

export const healthSet = (health?: string) =>
  new Set(health ? health.split(',').filter(Boolean) : [])

export const plural = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

export const countText = (
  counts: TAppFleetCounts | undefined,
  active: boolean
) => {
  const total = counts?.total ?? 0
  if (active) return `${counts?.matched ?? 0} of ${total} match`
  return plural(total, 'install')
}

export const rolloutSegments = (counts?: TAppFleetCounts) =>
  (['applied', 'behind', 'unknown'] as const)
    .map((state) => ({ state, count: counts?.[state] ?? 0 }))
    .filter(({ count }) => count > 0)
