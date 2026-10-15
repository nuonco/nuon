import { BranchOverview } from '@/components/branches/BranchOverview'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { branchOverviewVariant, useBranchOverviewV2 } from '@/hooks/use-branch-overview-v2'
import { BranchDetail } from '../BranchDetail'

export const BranchOverviewTab = () => {
  const variant = branchOverviewVariant(
    useNewAppIA(),
    useBranchOverviewV2()
  )

  if (variant === 'legacy') return <BranchDetail />
  return <BranchOverview variant={variant} />
}
