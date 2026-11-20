import type { ReactNode } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import {
  getAppFleet,
  getAppLabels,
  getComponents,
  toLabelColorMap,
} from '@/lib'
import type { TAppFleetFilters } from '@/types'
import { AppFleet } from './AppFleet'
import { FleetGroupInstallsContainer } from './FleetGroupInstallsContainer'
import {
  FLEET_FILTER_KEYS,
  hasActiveFilters,
  parseFleetFilters,
  parseFleetView,
  type TFleetView,
} from './fleet-utils'

export const AppFleetContainer = ({
  emptyAction,
  pollInterval = 20000,
  shouldPoll = true,
}: {
  emptyAction?: ReactNode
  pollInterval?: number
  shouldPoll?: boolean
}) => {
  const { org } = useOrg()
  const { app } = useApp()
  const [searchParams, setSearchParams] = useSearchParams()
  const filters = parseFleetFilters(searchParams)
  const view = parseFleetView(searchParams)

  const { data: fleet, isLoading } = useQuery({
    queryKey: ['app-fleet', org?.id, app?.id, filters],
    queryFn: () => getAppFleet({ orgId: org!.id, appId: app!.id, ...filters }),
    enabled: !!org?.id && !!app?.id,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll ? pollInterval : false,
  })

  const { data: appLabels } = useQuery({
    queryKey: ['app-labels', org?.id, app?.id],
    queryFn: () => getAppLabels({ orgId: org!.id, appId: app!.id }),
    enabled: !!org?.id && !!app?.id,
  })
  const labelColors = toLabelColorMap(appLabels)

  const { data: components } = useQuery({
    queryKey: ['app-fleet-components', org?.id, app?.id],
    queryFn: () =>
      getComponents({ orgId: org!.id, appId: app!.id, limit: 100 }),
    enabled: !!org?.id && !!app?.id,
  })
  const componentChoices = (components?.data ?? [])
    .filter((component) => component.id)
    .map((component) => ({
      value: component.id!,
      label: component.name ?? component.id!,
    }))
    .sort((a, b) => a.label.localeCompare(b.label))

  const updateParams = (update: (params: URLSearchParams) => void) =>
    setSearchParams(
      (prev) => {
        const params = new URLSearchParams(prev)
        update(params)
        return params
      },
      { replace: true }
    )

  const onFiltersChange = (next: Partial<TAppFleetFilters>) =>
    updateParams((params) => {
      for (const [key, value] of Object.entries(next)) {
        if (value) params.set(key, value)
        else params.delete(key)
      }
    })

  const onClearFilters = () =>
    updateParams((params) => {
      for (const key of FLEET_FILTER_KEYS) params.delete(key)
    })

  const onViewChange = (next: TFleetView) =>
    updateParams((params) => {
      if (next === 'board') params.set('view', next)
      else params.delete('view')
    })

  return (
    <AppFleet
      fleet={fleet}
      orgId={org?.id}
      appId={app?.id}
      isLoading={isLoading}
      filters={filters}
      isFiltered={hasActiveFilters(filters)}
      onFiltersChange={onFiltersChange}
      onClearFilters={onClearFilters}
      view={view}
      onViewChange={onViewChange}
      labelColors={labelColors}
      emptyAction={emptyAction}
      componentChoices={componentChoices}
      labelOptions={(appLabels?.labels ?? [])
        .filter((label) => label.entity_types?.includes('install'))
        .flatMap((label) =>
          (label.values ?? []).map((value) => `${label.key}:${value}`)
        )
        .sort()}
      renderGroupInstalls={(branch, group) => (
        <FleetGroupInstallsContainer
          appId={app?.id ?? ''}
          branchId={branch.id}
          groupName={group.name}
          filters={filters}
          labelColors={labelColors}
        />
      )}
    />
  )
}
