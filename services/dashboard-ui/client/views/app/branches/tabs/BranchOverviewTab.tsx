import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { BranchDetail } from '../BranchDetail'
import { BranchRunsTab } from './BranchRunsTab'

export const BranchOverviewTab = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchRunsTab /> : <BranchDetail />
}
