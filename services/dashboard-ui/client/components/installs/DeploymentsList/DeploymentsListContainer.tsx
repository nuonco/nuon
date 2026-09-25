import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRefreshErrorToast } from '@/hooks/use-refresh-error-toast'
import { getInstallDeployments } from '@/lib'
import { useSSETimelineQuery } from '@/lib/sse/use-sse-timeline-query'
import type { TInstallDeploymentRecordType } from '@/types'
import {
  datePresetQueryParameter,
  statusFilterParameter,
  WORKFLOW_DATE_LABELS,
  workflowStatusOptions,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import {
  DEPLOYMENT_TYPE_LABELS,
  DeploymentsListPresenter,
  type IDeploymentFilter,
} from './DeploymentsListPresenter'

const PAGE_LIMIT = 20
const SEARCH_DEBOUNCE_MS = 300
const FILTER_PARAMS = ['search', 'status', 'type', 'resource', 'since']

const readSet = <T extends string>(
  searchParams: URLSearchParams,
  key: string,
  options: readonly T[]
) => {
  const valid = new Set<string>(options)
  return new Set(
    (searchParams.get(key)?.split(',') ?? []).filter((value): value is T =>
      valid.has(value)
    )
  )
}

export interface IDeploymentsListContainer {
  pollInterval?: number
  shouldPoll?: boolean
}

export const DeploymentsListContainer = ({
  pollInterval = 20000,
  shouldPoll = false,
}: IDeploymentsListContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const [searchParams, setSearchParams] = useSearchParams()

  const offset = Number(searchParams.get('offset') ?? 0)
  const since = searchParams.get('since')
  const filter: IDeploymentFilter = {
    search: searchParams.get('search') ?? '',
    status: readSet<TWorkflowStatusOption>(
      searchParams,
      'status',
      workflowStatusOptions()
    ),
    type: readSet(
      searchParams,
      'type',
      Object.keys(DEPLOYMENT_TYPE_LABELS) as TInstallDeploymentRecordType[]
    ),
    resource: searchParams.get('resource') ?? undefined,
    date:
      since && since in WORKFLOW_DATE_LABELS
        ? (since as TWorkflowDatePreset)
        : undefined,
  }
  const status = statusFilterParameter(filter.status)
  const type = filter.type.size ? [...filter.type].join(',') : undefined
  const createdAtGte = useMemo(() => datePresetQueryParameter(since), [since])
  const queryKey = [
    'install-deployments',
    org?.id,
    install?.id,
    offset,
    status,
    type,
    filter.resource,
    createdAtGte,
    filter.search,
  ]
  const sseUrl = useMemo(() => {
    if (!org?.id || !install?.id) return undefined

    const params = new URLSearchParams({
      limit: String(PAGE_LIMIT),
      offset: String(offset),
    })
    if (status) params.set('status', status)
    if (type) params.set('type', type)
    if (filter.resource) params.set('resource', filter.resource)
    if (createdAtGte) params.set('created_at_gte', createdAtGte)
    if (filter.search) params.set('search', filter.search)
    return `/api/orgs/${org.id}/installs/${install.id}/deployments/sse?${params}`
  }, [
    createdAtGte,
    filter.resource,
    filter.search,
    install?.id,
    offset,
    org?.id,
    status,
    type,
  ])

  const [search, setSearch] = useState(filter.search)
  const searchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    setSearch(filter.search)
  }, [filter.search])

  useEffect(
    () => () => {
      if (searchTimer.current) clearTimeout(searchTimer.current)
    },
    []
  )

  const onRefreshError = useRefreshErrorToast()
  const { data, isLoading, error } = useSSETimelineQuery({
    sseUrl,
    queryKey,
    queryFn: () =>
      getInstallDeployments({
        orgId: org!.id,
        installId: install!.id,
        offset,
        limit: PAGE_LIMIT,
        status,
        type,
        resource: filter.resource,
        createdAtGte,
        search: filter.search || undefined,
      }),
    enabled: !!org?.id && !!install?.id,
    shouldPoll,
    pollInterval,
    eventName: 'deployments',
    onError: onRefreshError,
  })

  const writeParam = (key: string, value?: string) => {
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        if (value) next.set(key, value)
        else next.delete(key)
        next.delete('offset')
        return next
      },
      { replace: true }
    )
  }

  const writeSet = (key: string, selected: Set<string>) =>
    writeParam(key, selected.size ? [...selected].join(',') : undefined)

  const handleSearchChange = (value: string) => {
    setSearch(value)
    if (searchTimer.current) clearTimeout(searchTimer.current)
    searchTimer.current = setTimeout(
      () => writeParam('search', value),
      SEARCH_DEBOUNCE_MS
    )
  }

  const handleClearFilters = () => {
    if (searchTimer.current) clearTimeout(searchTimer.current)
    setSearch('')
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        FILTER_PARAMS.forEach((key) => next.delete(key))
        next.delete('offset')
        return next
      },
      { replace: true }
    )
  }

  return (
    <DeploymentsListPresenter
      deployments={data?.deployments ?? []}
      isLoading={isLoading}
      error={error}
      pagination={{
        hasNext: data?.has_more ?? false,
        offset,
        limit: PAGE_LIMIT,
      }}
      orgId={org?.id ?? ''}
      appId={install?.app_id ?? ''}
      installId={install?.id ?? ''}
      search={search}
      filter={filter}
      onSearchChange={handleSearchChange}
      onStatusChange={(value) => writeSet('status', value)}
      onTypeChange={(value) => writeSet('type', value)}
      onResourceChange={(value) => writeParam('resource', value)}
      onDateChange={(value) => writeParam('since', value)}
      onClearFilters={handleClearFilters}
    />
  )
}
