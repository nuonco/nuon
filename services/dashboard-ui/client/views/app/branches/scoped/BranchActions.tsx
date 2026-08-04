import { useMemo } from 'react'
import { ActionsTable } from '@/components/actions/ActionsTable'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { latestBranchConfig } from '@/utils/branch-utils'
import { Actions } from '../../Actions'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchActionsContent = () => {
  const { branch } = useBranch()
  const filterIds = useMemo(
    () => latestBranchConfig(branch)?.action_ids ?? [],
    [branch]
  )

  return (
    <BranchTabPage
      tab="Actions"
      heading="Actions"
      subheading="The day-2 operations defined by this branch's configuration."
    >
      <ActionsTable filterIds={filterIds} />
    </BranchTabPage>
  )
}

export const BranchActions = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchActionsContent /> : <Actions />
}
