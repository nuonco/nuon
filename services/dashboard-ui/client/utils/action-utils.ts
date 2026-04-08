import type { TActionConfig, TInstallActionRun } from '@/types'

export function sortByIdx<T extends { idx?: number }>(items: T[]): T[] {
  return items.slice().sort((a, b) => {
    if (a.idx === undefined && b.idx === undefined) return 0
    if (a.idx === undefined) return -1
    if (b.idx === undefined) return 1
    return a.idx - b.idx
  })
}

export type THydratedActionRunSteps = Array<
  TInstallActionRun['steps'][number] & { name?: string; idx?: number }
>

/**
 * Hydrates action run steps with their corresponding idx and name from the config.
 * @param params Object with steps and stepConfigs
 * @returns Array of hydrated steps.
 */
export function hydrateActionRunSteps({
  steps,
  stepConfigs,
}: {
  steps: TInstallActionRun['steps']
  stepConfigs: TActionConfig['steps']
}): THydratedActionRunSteps {
  if (!steps) return []

  // Build a map for fast lookup by step id
  const configMap = (stepConfigs ?? []).reduce<
    Record<string, { name?: string; idx?: number }>
  >((acc, cfg) => {
    acc[cfg.id] = { name: cfg.name, idx: cfg.idx }
    return acc
  }, {})

  return steps.map((step, index) => {
    const config = configMap[String(step.step_id)]
    const adhocName = (step as any).adhoc_config?.name
    return {
      ...step,
      ...config,
      name: config?.name ?? adhocName,
      idx: config?.idx ?? index,
    }
  })
}
