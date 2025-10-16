import type { TActionConfig, TInstallActionRun } from '@/types'

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
  if (!steps || !stepConfigs) return steps ?? []

  // Build a map for fast lookup by step id
  const configMap = stepConfigs.reduce<
    Record<string, { name?: string; idx?: number }>
  >((acc, cfg) => {
    acc[cfg.id] = { name: cfg.name, idx: cfg.idx }
    return acc
  }, {})

  return steps.map((step) => ({
    ...step,
    ...configMap[step.step_id], // name and idx from config
  }))
}
