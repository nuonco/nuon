import type { TAppBranch, TAppBranchConfig } from '@/types'

export const latestBranchConfig = (
  branch: TAppBranch
): TAppBranchConfig | undefined => {
  if (!branch.configs?.length) return undefined
  return [...branch.configs].sort(
    (a, b) => (b.config_number || 0) - (a.config_number || 0)
  )[0]
}
