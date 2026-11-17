export default {
  title: 'Playground / Fleet v2',
}

import { useState } from 'react'
import { HealthFilterDropdown } from './HealthFilterDropdown'
import { InstallFleet } from './InstallFleet'
import { LabelsFilterDropdown } from './LabelsFilterDropdown'
import {
  largeFleet,
  largeFleetBranches,
  largeFleetConfigs,
  largeFleetOverviews,
  largeFleetRuns,
  largeFleetStatus,
  mixedFleet,
  mixedFleetBranches,
  mixedFleetConfigs,
  mixedFleetOverviews,
  mixedFleetRuns,
  mixedFleetStatus,
} from './fixtures'

const mixedProps = {
  installs: mixedFleet,
  branches: mixedFleetBranches,
  branchRuns: mixedFleetRuns,
  branchConfigs: mixedFleetConfigs,
  statusById: mixedFleetStatus,
  overviewById: mixedFleetOverviews,
}

const largeProps = {
  installs: largeFleet,
  branches: largeFleetBranches,
  branchRuns: largeFleetRuns,
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
  const [selected, setSelected] = useState(new Set(['unhealthy', 'degraded']))
  return (
    <div className="flex justify-end">
      <HealthFilterDropdown selected={selected} onChange={setSelected} />
    </div>
  )
}

export const LabelsFilter = () => {
  const [selected, setSelected] = useState(new Set(['env:prod']))
  return (
    <div className="flex justify-end">
      <LabelsFilterDropdown
        options={['env:prod', 'env:stage', 'region:us-east-1', 'tier:canary']}
        selected={selected}
        onChange={setSelected}
      />
    </div>
  )
}

export const Loading = () => <InstallFleet installs={[]} isLoading />

export const Empty = () => <InstallFleet installs={[]} />
