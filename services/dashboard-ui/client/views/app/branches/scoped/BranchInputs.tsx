import { BranchInputs as BranchInputsContent } from '@/components/branches/BranchInputs'
import { useNewAppIA } from '@/hooks/use-new-app-ia'
import { Overview } from '../../Overview'
import { BranchTabPage } from '../tabs/BranchTabPage'

const NewBranchInputs = () => (
  <BranchTabPage
    tab="Inputs"
    heading="Inputs"
    subheading="The inputs defined by this branch's configuration."
  >
    <BranchInputsContent />
  </BranchTabPage>
)

export const BranchInputs = () => {
  const hasNewAppIA = useNewAppIA()

  return hasNewAppIA ? <NewBranchInputs /> : <Overview />
}
