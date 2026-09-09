import type { ReactNode } from 'react'
import { keepPreviousData, useQueries, useQuery } from '@tanstack/react-query'
import {
  getAppLabels,
  getBranches,
  getInstallLabelKeys,
  getInstalls,
  toLabelColorMap,
} from '@/lib'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { commaSetQueryParameter } from '../../../utils/list-query'
import { useOrg } from '../../../providers/org-provider'
import { Badge } from '../../atoms/Badge'
import {
  InstallsTable,
  type IInstallFilter,
  type TLabelColors,
} from './InstallsTable'

const PAGE_SIZE = 20
const INSTALL_FILTERS = {
  labels: commaSetQueryParameter('labels'),
  branches: commaSetQueryParameter('branches'),
}

const filterControl = ({
  label,
  values,
  constrained,
  onChange,
  renderOption,
}: {
  label: string
  values: string[]
  constrained: Set<string>
  onChange: (value: Set<string>) => void
  renderOption?: (value: string) => ReactNode
}): IInstallFilter | undefined => {
  const allValues = [...new Set([...values, ...constrained])].sort()
  if (!allValues.length) return undefined

  return {
    label,
    options: allValues.map((value) => ({
      value,
      label: renderOption?.(value) ?? value,
      textValue: value,
    })),
    selected: constrained,
    constrained: constrained.size > 0,
    onToggle: (value) => {
      const next = new Set(constrained)
      if (next.has(value)) next.delete(value)
      else next.add(value)
      onChange(next)
    },
    onIsolate: (value) =>
      onChange(
        constrained.size === 1 && constrained.has(value)
          ? new Set()
          : new Set([value])
      ),
    onReset: () => onChange(new Set()),
  }
}

export const InstallsTableContainer = () => {
  const { orgId } = useOrg()
  const list = useListQueryState({
    pageSize: PAGE_SIZE,
    filters: INSTALL_FILTERS,
  })

  const {
    data: result,
    isLoading,
    isPlaceholderData,
    error,
  } = useQuery({
    queryKey: ['installs', orgId, ...list.queryKey],
    queryFn: () =>
      getInstalls({
        orgId: orgId!,
        q: list.search || undefined,
        offset: list.offset,
        limit: list.pageSize,
        labels: [...list.filters.labels].join(',') || undefined,
        branches: [...list.filters.branches].join(',') || undefined,
      }),
    enabled: !!orgId,
    placeholderData: keepPreviousData,
    refetchInterval: 20_000,
  })

  const { data: labelValues } = useQuery({
    queryKey: ['install-label-keys', orgId],
    queryFn: () => getInstallLabelKeys({ orgId: orgId! }),
    enabled: !!orgId,
    staleTime: 60_000,
  })

  const { data: branchesResult } = useQuery({
    queryKey: ['org-branch-names', orgId],
    queryFn: () => getBranches({ orgId: orgId!, limit: 100 }),
    enabled: !!orgId,
    staleTime: 60_000,
  })

  const installs = result?.data ?? []
  const appIds = [
    ...new Set(
      installs.map((install) => install?.app_id).filter((id): id is string => !!id)
    ),
  ]

  const appLabelQueries = useQueries({
    queries: appIds.map((appId) => ({
      queryKey: ['app-labels', orgId, appId],
      queryFn: () => getAppLabels({ orgId: orgId!, appId }),
      enabled: !!orgId,
      staleTime: 60_000,
    })),
  })

  const labelColors: TLabelColors = Object.fromEntries(
    appIds.map((appId, index) => [
      appId,
      toLabelColorMap(appLabelQueries[index]?.data),
    ])
  )

  const filterLabelColors: Record<string, string> = {}
  for (const query of appLabelQueries) {
    for (const label of query.data?.labels ?? []) {
      if (label.is_override || !filterLabelColors[label.key]) {
        filterLabelColors[label.key] = label.color
      }
    }
  }

  const labelOptions = Object.entries(labelValues ?? {}).flatMap(
    ([key, values]) => values.map((value) => `${key}:${value}`)
  )
  const branchNames = [
    ...new Set(
      (branchesResult?.data ?? [])
        .map((branch) => branch?.name)
        .filter((name): name is string => !!name)
    ),
  ]
  return (
    <InstallsTable
      installs={installs}
      orgId={orgId ?? ''}
      search={list.search}
      onSearchChange={list.setSearch}
      offset={list.offset}
      pageSize={list.pageSize}
      hasNext={result?.pagination?.hasNext ?? false}
      onOffsetChange={list.setOffset}
      labelFilter={filterControl({
        label: 'Labels',
        values: labelOptions,
        constrained: list.filters.labels,
        onChange: (value) => list.setFilter('labels', value),
        renderOption: (value) => {
          const [key, ...rest] = value.split(':')
          return (
            <Badge
              variant="code"
              labelKey={key}
              labelValue={rest.join(':')}
              color={filterLabelColors[key]}
            />
          )
        },
      })}
      branchFilter={filterControl({
        label: 'Branches',
        values: ['__none__', ...branchNames],
        constrained: list.filters.branches,
        onChange: (value) => list.setFilter('branches', value),
        renderOption: (value) =>
          value === '__none__' ? 'No branch' : value,
      })}
      labelColors={labelColors}
      loading={isLoading}
      fetching={isPlaceholderData}
      error={error}
    />
  )
}
