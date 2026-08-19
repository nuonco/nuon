import { useMemo } from 'react'
import { RunbooksTable } from '@/components/runbooks/RunbooksTable'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { latestBranchConfig } from '@/utils/branch-utils'
import { Runbooks } from '../../Runbooks'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchRunbooksContent = () => {
  const { branch } = useBranch()
  const filterIds = useMemo(
    () => latestBranchConfig(branch)?.runbook_ids ?? [],
    [branch]
  )

  return (
    <BranchTabPage
      tab="Runbooks"
      heading="Runbooks"
      subheading="The runbooks defined by this branch's configuration."
    >
      <RunbooksTable filterIds={filterIds} branchId={branch?.id} />
    </BranchTabPage>
  )
}

export const BranchRunbooks = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchRunbooksContent /> : <Runbooks />
}
