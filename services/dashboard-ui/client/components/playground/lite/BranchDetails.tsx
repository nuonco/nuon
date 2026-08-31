import { BranchLabels } from './BranchLabels'
import { InstallInputs } from './InstallInputs'
import { SourceCard } from './SourceCard'

export const BranchDetails = () => (
  <div className="flex flex-col gap-6">
    <SourceCard />
    <InstallInputs />
    <BranchLabels />
  </div>
)
