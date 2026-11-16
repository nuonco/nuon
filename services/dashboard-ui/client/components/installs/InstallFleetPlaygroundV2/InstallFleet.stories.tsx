export default {
  title: 'Playground / Fleet v2',
}

import { useState } from 'react'
import { HealthFilterDropdown } from './HealthFilterDropdown'
import { InstallFleet } from './InstallFleet'
import { mixedFleet, mixedFleetConfigs, mixedFleetStatus } from './fixtures'

export const MixedFleet = () => (
  <InstallFleet
    installs={mixedFleet}
    branchConfigs={mixedFleetConfigs}
    statusById={mixedFleetStatus}
  />
)

export const Board = () => (
  <InstallFleet
    installs={mixedFleet}
    branchConfigs={mixedFleetConfigs}
    statusById={mixedFleetStatus}
    defaultView="board"
  />
)

export const HealthFilter = () => {
  const [selected, setSelected] = useState(
    new Set(['unhealthy', 'degraded'])
  )
  return (
    <div className="flex justify-end">
      <HealthFilterDropdown selected={selected} onChange={setSelected} />
    </div>
  )
}

export const Loading = () => <InstallFleet installs={[]} isLoading />

export const Empty = () => <InstallFleet installs={[]} />
