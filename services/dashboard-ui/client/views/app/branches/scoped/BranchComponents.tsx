import { useMemo } from 'react'
import { ComponentsTable } from '@/components/components/ComponentsTable'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { latestBranchConfig } from '@/utils/branch-utils'
import { Components } from '../../Components'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchComponentsContent = () => {
  const { branch } = useBranch()
  const filterIds = useMemo(
    () => latestBranchConfig(branch)?.component_ids ?? [],
    [branch]
  )

  return (
    <BranchTabPage
      tab="Components"
      heading="Components"
      subheading="The components defined by this branch's configuration."
    >
      <div className="flex flex-auto min-w-0">
        <ComponentsTable filterIds={filterIds} />
      </div>
    </BranchTabPage>
  )
}

export const BranchComponents = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchComponentsContent /> : <Components />
}
