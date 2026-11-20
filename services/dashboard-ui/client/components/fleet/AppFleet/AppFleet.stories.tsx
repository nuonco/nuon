import { useState } from 'react'
import type { TAppFleet, TAppFleetFilters } from '@/types'
import { AppFleet } from './AppFleet'
import { appFleet, emptyAppFleet, filteredAppFleet } from './AppFleet.fixtures'
import { FleetGroupInstalls } from './FleetGroupInstalls'
import { hasActiveFilters, type TFleetView } from './fleet-utils'

export default {
  title: 'Features / Fleet / App fleet',
}

const FleetStory = ({
  fleet,
  isLoading,
  initialView = 'table',
  initialFilters = {},
}: {
  fleet?: TAppFleet
  isLoading?: boolean
  initialView?: TFleetView
  initialFilters?: TAppFleetFilters
}) => {
  const [view, setView] = useState<TFleetView>(initialView)
  const [filters, setFilters] = useState<TAppFleetFilters>(initialFilters)
  return (
    <AppFleet
      fleet={fleet}
      orgId="org-1"
      appId="app-1"
      isLoading={isLoading}
      filters={filters}
      isFiltered={hasActiveFilters(filters)}
      onFiltersChange={(next) =>
        setFilters((current) => ({ ...current, ...next }))
      }
      onClearFilters={() => setFilters({})}
      view={view}
      onViewChange={setView}
      labelOptions={['env:prod', 'region:us-east-1', 'tier:canary']}
      componentChoices={[
        { value: 'cmp-api', label: 'api' },
        { value: 'cmp-worker', label: 'worker' },
      ]}
      renderGroupInstalls={(_, group) => (
        <FleetGroupInstalls
          installs={group.installs ?? []}
          orgId="org-1"
          appId="app-1"
          onLoadMore={group.has_more ? () => undefined : undefined}
        />
      )}
    />
  )
}

export const Table = () => <FleetStory fleet={appFleet} />

export const Board = () => <FleetStory fleet={appFleet} initialView="board" />

export const Filtered = () => (
  <FleetStory
    fleet={filteredAppFleet}
    initialFilters={{ health: 'degraded' }}
  />
)

export const FilteredBoard = () => (
  <FleetStory
    fleet={filteredAppFleet}
    initialView="board"
    initialFilters={{ health: 'degraded' }}
  />
)

export const NoMatches = () => (
  <FleetStory
    fleet={{
      ...appFleet,
      summary: { ...appFleet.summary, matched: 0 },
    }}
    initialFilters={{ rollout: 'behind' }}
  />
)

export const Empty = () => <FleetStory fleet={emptyAppFleet} />

export const Loading = () => <FleetStory isLoading />
