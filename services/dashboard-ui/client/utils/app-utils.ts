import type { TAppConfig, TInstall } from '@/types'

export function hasNewerAppConfig(
  latestConfig: TAppConfig | undefined,
  install: TInstall | undefined,
): boolean {
  if (!latestConfig?.id || !install?.app_config_id) return false
  return latestConfig.id !== install.app_config_id
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
