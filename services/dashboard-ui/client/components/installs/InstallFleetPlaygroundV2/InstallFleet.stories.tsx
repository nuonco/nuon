export default {
  title: 'Playground / Fleet v2',
}

import { useState } from 'react'
import { HealthFilterDropdown } from './HealthFilterDropdown'
import { InstallFleet } from './InstallFleet'
import {
  largeFleet,
  largeFleetConfigs,
  largeFleetStatus,
  mixedFleet,
  mixedFleetConfigs,
  mixedFleetStatus,
} from './fixtures'

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

export const LargeFleet = () => (
  <InstallFleet
    installs={largeFleet}
    branchConfigs={largeFleetConfigs}
    statusById={largeFleetStatus}
  />
)

export const LargeFleetBoard = () => (
  <InstallFleet
    installs={largeFleet}
    branchConfigs={largeFleetConfigs}
    statusById={largeFleetStatus}
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
