import { useOrgFeatureFlag } from '@/hooks/use-org-feature-flag'

export type TBranchOverviewVariant = 'legacy' | 'current' | 'v2'

export const branchOverviewVariant = (
  newAppIA: boolean,
  overviewV2: boolean
): TBranchOverviewVariant => {
  if (!newAppIA) return 'legacy'
  return overviewV2 ? 'v2' : 'current'
}

export const useBranchOverviewV2 = () =>
  useOrgFeatureFlag('branch-overview-v2')
