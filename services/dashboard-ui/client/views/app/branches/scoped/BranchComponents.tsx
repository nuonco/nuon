import { ComponentsTable } from '@/components/components/ComponentsTable'
import { useBranch } from '@/hooks/use-branch'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { Components } from '../../Components'
import { BranchTabPage } from '../tabs/BranchTabPage'

const BranchComponentsContent = () => {
  const { branch } = useBranch()

  return (
    <BranchTabPage
      tab="Components"
      heading="Components"
      subheading="The components defined by this branch's configuration."
    >
      <div className="flex flex-auto min-w-0">
        <ComponentsTable branchId={branch?.id} />
      </div>
    </BranchTabPage>
  )
}

export const BranchComponents = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <BranchComponentsContent /> : <Components />
}
