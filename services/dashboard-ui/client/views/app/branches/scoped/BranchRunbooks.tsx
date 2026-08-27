import { RunbooksTable } from '@/components/runbooks/RunbooksTable'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { Runbooks } from '../../Runbooks'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchRunbooksContent = () => {
  const { branch } = useBranch()

  return (
    <BranchTabPage
      tab="Runbooks"
      heading="Runbooks"
      subheading="The runbooks defined by this branch's configuration."
    >
      <RunbooksTable branchId={branch?.id} />
    </BranchTabPage>
  )
}

export const BranchRunbooks = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchRunbooksContent /> : <Runbooks />
}
