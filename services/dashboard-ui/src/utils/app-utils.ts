import type { TAppConfig } from '@/types'

export function normalizeAppInputGroups(
  groups: TAppConfig['input']['input_groups'],
  inputs: TAppConfig['input']['inputs']
) {
  return groups.map((group) => ({
    ...group,
    app_inputs: inputs.filter((input) => input.group_id === group.id),
  }))
}
