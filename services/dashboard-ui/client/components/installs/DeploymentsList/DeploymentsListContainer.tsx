import { useEffect, useMemo, useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { useRefreshErrorToast } from '@/hooks/use-refresh-error-toast'
import { getAppBranch, getInstallDeploymentSummaries } from '@/lib'
import { latestBranchConfig } from '@/utils/branch-utils'
import { vcsRepo } from '@/utils/vcs-urls'
import { useSSETimelineQuery } from '@/lib/sse/use-sse-timeline-query'
import { buildQueryParams } from '@/utils/build-query-params'
import { ACTIVE_DEPLOYMENT_STATUSES } from '@/components/installs/DeploymentDetail/deployment-progress'
import {
  datePresetQueryParameter,
  statusFilterParameter,
  WORKFLOW_DATE_LABELS,
  workflowStatusOptions,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import {
  DEPLOYMENT_FILTER_TYPES,
  DeploymentsListPresenter,
  type IDeploymentFilter,
  type IDeploymentsListPresenter,
} from './DeploymentsListPresenter'

const PAGE_LIMIT = 20
const ACTIVE_PAGE_LIMIT = 4
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
  const branchId = install?.app_branch?.id
  const { data: branch } = useQuery({
    placeholderData: keepPreviousData,
    queryKey: ['install-branch-repo', org?.id, install?.app_id, branchId],
    queryFn: () =>
      getAppBranch({
        orgId: org!.id,
        appId: install!.app_id!,
        branchId: branchId!,
        latestConfig: true,
      }),
    enabled: !!org?.id && !!install?.app_id && !!branchId,
  })
  const repo = branch ? vcsRepo(latestBranchConfig(branch)) : undefined

  const since = searchParams.get('since')
  const filter: IDeploymentFilter = {
    search: searchParams.get('search') ?? '',
    status: readSet<TWorkflowStatusOption>(
      searchParams,
      'status',
      workflowStatusOptions()
    ),
    type: readSet(searchParams, 'type', DEPLOYMENT_FILTER_TYPES),
    resource: searchParams.get('resource') ?? undefined,
    date:
      since && since in WORKFLOW_DATE_LABELS
        ? (since as TWorkflowDatePreset)
        : undefined,
  }
  const status = filter.status.has('running')
    ? [statusFilterParameter(filter.status), 'approval-awaiting', 'approved']
        .filter(Boolean)
        .join(',')
    : statusFilterParameter(filter.status)
  const type = filter.type.size ? [...filter.type].join(',') : undefined
  const createdAtGte = useMemo(() => datePresetQueryParameter(since), [since])

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
    <DeploymentFeeds
      query={{
        orgId: org?.id ?? '',
        installId: install?.id ?? '',
        status,
        type,
        resource: filter.resource,
        createdAtGte,
        search: filter.search || undefined,
      }}
      shouldPoll={shouldPoll}
      pollInterval={pollInterval}
      presenterProps={{
        orgId: org?.id ?? '',
        appId: install?.app_id ?? '',
        installId: install?.id ?? '',
        repo,
        search,
        filter,
        onSearchChange: handleSearchChange,
        onStatusChange: (value) => writeSet('status', value),
        onTypeChange: (value) => writeSet('type', value),
        onResourceChange: (value) => writeParam('resource', value),
        onDateChange: (value) => writeParam('since', value),
        onClearFilters: handleClearFilters,
      }}
    />
  )
}

const useDeploymentSection = (
  query: Parameters<typeof getInstallDeploymentSummaries>[0],
  visible: boolean,
  shouldPoll: boolean,
  pollInterval: number
) => {
  const onError = useRefreshErrorToast()
  const sseUrl =
    visible && query.orgId && query.installId
      ? `/api/orgs/${query.orgId}/installs/${query.installId}/deployment-summaries/sse${buildQueryParams(
          {
            state: query.state,
            sort: query.sort,
            cursor: query.cursor,
            limit: query.limit,
            status: query.status,
            type: query.type,
            resource: query.resource,
            created_at_gte: query.createdAtGte,
            search: query.search,
          }
        )}`
      : undefined
  return useSSETimelineQuery({
    queryKey: ['install-deployments', query],
    queryFn: () => getInstallDeploymentSummaries(query),
    sseUrl,
    enabled: !!query.orgId && !!query.installId && visible,
    shouldPoll: shouldPoll && visible,
    pollInterval,
    eventName: 'deployments',
    retainPreviousData: true,
    onError,
  })
}

const useStableSection = <E,>(section: {
  data?: unknown
  error: E
  isLoading: boolean
}) => {
  const heldError = useRef<E | null>(null)
  if (section.data !== undefined) heldError.current = null
  else if (section.error) heldError.current = section.error
  const error =
    section.data === undefined
      ? (section.error ?? heldError.current)
      : section.error
  return {
    error,
    isLoading: section.isLoading && section.data === undefined && error == null,
  }
}

const DeploymentFeeds = ({
  query,
  presenterProps,
  shouldPoll,
  pollInterval,
}: {
  query: Parameters<typeof getInstallDeploymentSummaries>[0]
  presenterProps: Omit<
    IDeploymentsListPresenter,
    'deployments' | 'isLoading' | 'pagination'
  >
  shouldPoll: boolean
  pollInterval: number
}) => {
  const filterKey = JSON.stringify(query)
  const [previousFilterKey, setPreviousFilterKey] = useState(filterKey)
  const [activeLimit, setActiveLimit] = useState(ACTIVE_PAGE_LIMIT)
  const [historyCursors, setHistoryCursors] = useState<(string | undefined)[]>([
    undefined,
  ])
  if (filterKey !== previousFilterKey) {
    setPreviousFilterKey(filterKey)
    setActiveLimit(ACTIVE_PAGE_LIMIT)
    setHistoryCursors([undefined])
  }
  const statuses = query.status?.split(',')
  const showActive =
    !statuses ||
    statuses.some((status) => ACTIVE_DEPLOYMENT_STATUSES.has(status))
  const showHistory =
    !statuses ||
    statuses.some((status) => !ACTIVE_DEPLOYMENT_STATUSES.has(status))
  const active = useDeploymentSection(
    { ...query, state: 'active', sort: 'attention', limit: activeLimit },
    showActive,
    shouldPoll,
    pollInterval
  )
  const history = useDeploymentSection(
    {
      ...query,
      state: 'finished',
      cursor: historyCursors.at(-1),
      limit: PAGE_LIMIT,
    },
    showHistory,
    shouldPoll,
    pollInterval
  )
  const activeSection = useStableSection(active)
  const historySection = useStableSection(history)
  return (
    <DeploymentsListPresenter
      {...presenterProps}
      activeDeployments={active.data?.deployments ?? []}
      activeTotal={active.data?.total}
      activeLoading={activeSection.isLoading}
      activeError={activeSection.error}
      hasMoreActive={active.data?.has_more}
      showActive={showActive}
      showHistory={showHistory}
      onLoadMoreActive={() =>
        setActiveLimit((limit) => limit + ACTIVE_PAGE_LIMIT)
      }
      deployments={history.data?.deployments ?? []}
      isLoading={historySection.isLoading}
      error={historySection.error}
      pagination={{
        hasNext: !!history.data?.next_cursor && history.data.has_more,
        offset: (historyCursors.length - 1) * PAGE_LIMIT,
        limit: PAGE_LIMIT,
        onNext: () => {
          if (history.data?.next_cursor)
            setHistoryCursors((cursors) => [
              ...cursors,
              history.data!.next_cursor,
            ])
        },
        onPrevious: () =>
          setHistoryCursors((cursors) =>
            cursors.length > 1 ? cursors.slice(0, -1) : cursors
          ),
      }}
    />
  )
}
