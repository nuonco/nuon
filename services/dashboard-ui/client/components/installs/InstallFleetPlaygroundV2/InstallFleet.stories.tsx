export default {
  title: 'Playground / Fleet v2',
}

import { useState } from 'react'
import { HealthFilterDropdown } from './HealthFilterDropdown'
import { InstallFleet } from './InstallFleet'
import {
  largeFleet,
  largeFleetBranches,
  largeFleetConfigs,
  largeFleetOverviews,
  largeFleetStatus,
  mixedFleet,
  mixedFleetBranches,
  mixedFleetConfigs,
  mixedFleetOverviews,
  mixedFleetStatus,
} from './fixtures'

const mixedProps = {
  installs: mixedFleet,
  branches: mixedFleetBranches,
  branchConfigs: mixedFleetConfigs,
  statusById: mixedFleetStatus,
  overviewById: mixedFleetOverviews,
}

const largeProps = {
  installs: largeFleet,
  branches: largeFleetBranches,
  branchConfigs: largeFleetConfigs,
  statusById: largeFleetStatus,
  overviewById: largeFleetOverviews,
}

export const MixedFleet = () => <InstallFleet {...mixedProps} />

export const Board = () => <InstallFleet {...mixedProps} defaultView="board" />

export const LargeFleet = () => <InstallFleet {...largeProps} />

export const LargeFleetBoard = () => (
  <InstallFleet {...largeProps} defaultView="board" />
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
