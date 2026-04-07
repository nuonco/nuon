export default {
  title: 'Deploys/DeploySwitcher/DeploysSkeleton',
}

import { DeploysSkeleton } from './DeploysSkeleton'

export const Default = () => (
  <div className="flex flex-col gap-2 w-64">
    <DeploysSkeleton />
  </div>
)

export const CustomLimit = () => (
  <div className="flex flex-col gap-2 w-64">
    <DeploysSkeleton limit={3} />
  </div>
)
