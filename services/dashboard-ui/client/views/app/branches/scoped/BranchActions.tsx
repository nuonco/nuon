import { ActionsTable } from '@/components/actions/ActionsTable'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { Actions } from '../../Actions'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchActionsContent = () => {
  const { branch } = useBranch()

  return (
    <BranchTabPage
      tab="Actions"
      heading="Actions"
      subheading="The day-2 operations defined by this branch's configuration."
    >
      <ActionsTable branchId={branch?.id} />
    </BranchTabPage>
  )
}

export const BranchActions = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchActionsContent /> : <Actions />
}
