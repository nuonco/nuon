export default {
  title: 'Orgs/StatsGrid',
}

import { StatsGrid } from './StatsGrid'

export const Default = () => (
  <StatsGrid
    stats={[
      { label: 'Apps', value: 12 },
      { label: 'Installs', value: 48 },
      { label: 'Runners', value: 3 },
      { label: 'Active deploys', value: 7 },
    ]}
  />
)

export const TwoStats = () => (
  <StatsGrid
    stats={[
      { label: 'Total installs', value: 24 },
      { label: 'Active runners', value: 2 },
    ]}
  />
)

export const ZeroValues = () => (
  <StatsGrid
    stats={[
      { label: 'Apps', value: 0 },
      { label: 'Installs', value: 0 },
      { label: 'Runners', value: 0 },
      { label: 'Deploys', value: 0 },
    ]}
  />
)
