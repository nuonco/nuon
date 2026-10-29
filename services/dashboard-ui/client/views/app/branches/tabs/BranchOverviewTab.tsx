import { BranchOverview } from '@/components/branches/BranchOverview'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { BranchDetail } from '../BranchDetail'

export const BranchOverviewTab = () => {
  if (!useNewAppIA()) return <BranchDetail />
  return <BranchOverview />
}
