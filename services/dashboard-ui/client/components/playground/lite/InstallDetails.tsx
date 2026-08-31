import { BranchLabels } from './BranchLabels'
import { InstallInputs } from './InstallInputs'
import { StatTile } from './StatTile'

export const InstallDetails = () => (
  <div className="flex flex-col gap-6">
    <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
      {['Status', 'App', 'Branch', 'Created'].map((stat) => (
        <StatTile key={stat} label={stat} />
      ))}
    </div>

    <InstallInputs />
    <BranchLabels />
  </div>
)
