import type { TAppConfig, TInstall } from '@/types'

export function hasNewerAppConfig(
  latestConfig: TAppConfig | undefined,
  install: TInstall | undefined,
): boolean {
  if (!latestConfig?.id || !install?.app_config_id) return false
  return latestConfig.id !== install.app_config_id
}

export function hasStackConfigChanged(
  currentConfig: TAppConfig | undefined,
  latestConfig: TAppConfig | undefined,
): boolean {
  if (!currentConfig?.stack || !latestConfig?.stack) return false
  const a = currentConfig.stack
  const b = latestConfig.stack
  return (
    a.type !== b.type ||
    a.name !== b.name ||
    a.runner_nested_template_url !== b.runner_nested_template_url ||
    a.vpc_nested_template_url !== b.vpc_nested_template_url
  )
}

export function normalizeAppInputGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
