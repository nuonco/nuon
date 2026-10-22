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
import {
  ACTIVE_DEPLOYMENT_FILTERS,
  activeDeploymentCounts,
  type TActiveDeploymentFilter,
} from '@/components/installs/DeploymentDetail/deployment-progress'
import {
  datePresetQueryParameter,
  statusFilterParameter,
  WORKFLOW_DATE_LABELS,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import {
  DEPLOYMENT_FILTER_TYPES,
  DeploymentsListPresenter,
  HISTORY_STATUS_OPTIONS,
  type IDeploymentFilter,
  type IDeploymentsListPresenter,
} from './DeploymentsListPresenter'

const PAGE_LIMIT = 20
const ACTIVE_PAGE_LIMIT = 4
const ACTIVE_FETCH_LIMIT = 100
const SEARCH_DEBOUNCE_MS = 300
const FILTER_PARAMS = ['search', 'status', 'type', 'resource', 'since']
const ACTIVE_FILTER_VALUES = ACTIVE_DEPLOYMENT_FILTERS.map(
  (option) => option.value
)

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
  const installComponents = useMemo(
    () =>
      [
        ...new Set(
          (install?.install_components ?? [])
            .map((installComponent) => installComponent.component?.name)
            .filter((name): name is string => !!name)
        ),
      ].sort((a, b) => a.localeCompare(b)),
    [install?.install_components]
  )
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
      HISTORY_STATUS_OPTIONS
    ),
    type: readSet(searchParams, 'type', DEPLOYMENT_FILTER_TYPES),
    resource: searchParams.get('resource') ?? undefined,
    date:
      since && since in WORKFLOW_DATE_LABELS
        ? (since as TWorkflowDatePreset)
        : undefined,
  }
  const activeParam = searchParams.get('active')
  const activeFilter: TActiveDeploymentFilter =
    activeParam &&
    ACTIVE_FILTER_VALUES.includes(activeParam as TActiveDeploymentFilter)
      ? (activeParam as TActiveDeploymentFilter)
      : 'all'
  const status = statusFilterParameter(filter.status)
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
      activeQuery={{
        orgId: org?.id ?? '',
        installId: install?.id ?? '',
      }}
      historyQuery={{
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
        installComponents,
        search,
        filter,
        activeFilter,
        onActiveFilterChange: (value) =>
          setSearchParams(
            (current) => {
              const next = new URLSearchParams(current)
              if (value === 'all') next.delete('active')
              else next.set('active', value)
              return next
            },
            { replace: true }
          ),
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

type TDeploymentQuery = Parameters<typeof getInstallDeploymentSummaries>[0]

const useDeploymentSection = (
  query: TDeploymentQuery,
  shouldPoll: boolean,
  pollInterval: number
) => {
  const onError = useRefreshErrorToast()
  const sseUrl =
    query.orgId && query.installId
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
    enabled: !!query.orgId && !!query.installId,
    shouldPoll,
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
  activeQuery,
  historyQuery,
  presenterProps,
  shouldPoll,
  pollInterval,
}: {
  activeQuery: TDeploymentQuery
  historyQuery: TDeploymentQuery
  presenterProps: Omit<
    IDeploymentsListPresenter,
    'deployments' | 'isLoading' | 'pagination'
  >
  shouldPoll: boolean
  pollInterval: number
}) => {
  const activeKey = JSON.stringify(activeQuery)
  const historyKey = JSON.stringify(historyQuery)
  const [previousActiveKey, setPreviousActiveKey] = useState(activeKey)
  const [previousHistoryKey, setPreviousHistoryKey] = useState(historyKey)
  const [activeLimit, setActiveLimit] = useState(ACTIVE_PAGE_LIMIT)
  const [historyCursors, setHistoryCursors] = useState<(string | undefined)[]>([
    undefined,
  ])
  if (activeKey !== previousActiveKey) {
    setPreviousActiveKey(activeKey)
    setActiveLimit(ACTIVE_PAGE_LIMIT)
  }
  if (historyKey !== previousHistoryKey) {
    setPreviousHistoryKey(historyKey)
    setHistoryCursors([undefined])
  }
  const active = useDeploymentSection(
    {
      ...activeQuery,
      state: 'active',
      sort: 'attention',
      limit: Math.max(ACTIVE_FETCH_LIMIT, activeLimit),
    },
    shouldPoll,
    pollInterval
  )
  const history = useDeploymentSection(
    {
      ...historyQuery,
      state: 'finished',
      cursor: historyCursors.at(-1),
      limit: PAGE_LIMIT,
    },
    shouldPoll,
    pollInterval
  )
  const activeSection = useStableSection(active)
  const historySection = useStableSection(history)
  const loaded = active.data?.deployments ?? []
  const activeFilter = presenterProps.activeFilter ?? 'all'
  const statuses = ACTIVE_DEPLOYMENT_FILTERS.find(
    (option) => option.value === activeFilter
  )?.statuses
  const filtered = statuses
    ? loaded.filter((deployment) => statuses.includes(deployment.status))
    : loaded
  const visible = filtered.slice(0, activeLimit)
  return (
    <DeploymentsListPresenter
      {...presenterProps}
      activeDeployments={visible}
      activeCounts={
        active.data && !active.data.has_more
          ? activeDeploymentCounts(loaded)
          : undefined
      }
      activeTotal={
        activeFilter === 'all' && active.data?.total != null
          ? active.data.total
          : filtered.length
      }
      activeLoading={activeSection.isLoading}
      activeError={activeSection.error}
      hasMoreActive={
        visible.length < filtered.length ||
        (activeFilter === 'all' && !!active.data?.has_more)
      }
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
