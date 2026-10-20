import { useEffect, useMemo, useRef, useState } from 'react'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { useSearchParams } from 'react-router'
import { useAuth } from '@/hooks/use-auth'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallActivity } from '@/lib'
import type { TAPIError, TInstallActivityType } from '@/types'
import {
  datePresetQueryParameter,
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_GROUPS,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import {
  ACTIVITY_STATUS_OPTIONS,
  ACTIVITY_TRIGGER_LABELS,
  ACTIVITY_TYPE_LABELS,
  ActivityListPresenter,
  type IActivityFilter,
  type TActivityTriggerFilter,
} from './ActivityListPresenter'

const PAGE_LIMIT = 20
const SEARCH_DEBOUNCE_MS = 300
const SEARCH_PARAM = 'activity_search'
const STATUS_PARAM = 'activity_status'
const TYPE_PARAM = 'activity_type'
const TRIGGER_PARAM = 'activity_trigger'
const SINCE_PARAM = 'activity_since'
const OFFSET_PARAM = 'activity_offset'
const FILTER_PARAMS = [
  SEARCH_PARAM,
  STATUS_PARAM,
  TYPE_PARAM,
  TRIGGER_PARAM,
  SINCE_PARAM,
]

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

const activityStatusParameter = (
  selected: Iterable<TWorkflowStatusOption>
): string | undefined => {
  const statuses = new Set(
    [...selected].flatMap((option) => {
      const values = [...(WORKFLOW_STATUS_GROUPS[option] ?? [])]
      if (option === 'succeeded') values.push('finished')
      if (option === 'failed') values.push('timed-out')
      if (option === 'running') values.push('retried')
      return values
    })
  )
  return statuses.size ? [...statuses].join(',') : undefined
}

export interface IActivityListContainer {
  pollInterval?: number
  shouldPoll?: boolean
}

export const ActivityListContainer = ({
  pollInterval = 20000,
  shouldPoll = false,
}: IActivityListContainer) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { user } = useAuth()
  const [searchParams, setSearchParams] = useSearchParams()

  const since = searchParams.get(SINCE_PARAM)
  const filter: IActivityFilter = {
    search: searchParams.get(SEARCH_PARAM) ?? '',
    status: readSet<TWorkflowStatusOption>(
      searchParams,
      STATUS_PARAM,
      ACTIVITY_STATUS_OPTIONS
    ),
    type: readSet(
      searchParams,
      TYPE_PARAM,
      Object.keys(ACTIVITY_TYPE_LABELS) as TInstallActivityType[]
    ),
    trigger: readSet(
      searchParams,
      TRIGGER_PARAM,
      Object.keys(ACTIVITY_TRIGGER_LABELS) as TActivityTriggerFilter[]
    ),
    date:
      since && since in WORKFLOW_DATE_LABELS
        ? (since as TWorkflowDatePreset)
        : undefined,
  }
  const status = activityStatusParameter(filter.status)
  const type = filter.type.size ? [...filter.type].join(',') : undefined
  const actorId = filter.trigger.has('mine') ? user?.sub : undefined
  const excludeCron = filter.trigger.has('hide_cron')
  const createdAtGte = useMemo(() => datePresetQueryParameter(since), [since])
  const filterKey = JSON.stringify([
    'install-activity',
    org?.id,
    install?.id,
    status,
    type,
    actorId,
    excludeCron,
    createdAtGte,
    filter.search,
  ])
  const [previousFilterKey, setPreviousFilterKey] = useState(filterKey)
  const [cursors, setCursors] = useState<(string | undefined)[]>([undefined])
  if (previousFilterKey !== filterKey) {
    setPreviousFilterKey(filterKey)
    setCursors([undefined])
  }
  const cursor = cursors.at(-1)
  const offset = (cursors.length - 1) * PAGE_LIMIT

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

  const { data, isLoading, isPlaceholderData, error } = useQuery({
    queryKey: [filterKey, cursor],
    queryFn: () =>
      getInstallActivity({
        orgId: org!.id,
        installId: install!.id,
        cursor,
        limit: PAGE_LIMIT,
        status,
        type,
        actorId,
        excludeCron,
        createdAtGte,
        search: filter.search || undefined,
      }),
    enabled: !!org?.id && !!install?.id,
    placeholderData: keepPreviousData,
    refetchInterval: shouldPoll && !cursor ? pollInterval : false,
  })

  const writeParam = (key: string, value?: string) => {
    setSearchParams(
      (current) => {
        const next = new URLSearchParams(current)
        if (value) next.set(key, value)
        else next.delete(key)
        next.delete(OFFSET_PARAM)
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
      () => writeParam(SEARCH_PARAM, value),
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
        next.delete(OFFSET_PARAM)
        return next
      },
      { replace: true }
    )
  }

  return (
    <ActivityListPresenter
      activity={data?.activity ?? []}
      isLoading={isLoading || isPlaceholderData}
      error={(error as TAPIError | null) ?? null}
      pagination={{
        hasNext: !!data?.next_cursor && !!data?.has_more,
        offset,
        limit: PAGE_LIMIT,
        onNext: () => {
          if (data?.next_cursor && !isPlaceholderData)
            setCursors((current) => [...current, data.next_cursor])
        },
        onPrevious: () =>
          setCursors((current) =>
            current.length > 1 ? current.slice(0, -1) : current
          ),
      }}
      orgId={org?.id ?? ''}
      installId={install?.id ?? ''}
      search={search}
      filter={filter}
      onSearchChange={handleSearchChange}
      onStatusChange={(value) => writeSet(STATUS_PARAM, value)}
      onTypeChange={(value) => writeSet(TYPE_PARAM, value)}
      onTriggerChange={(value) => writeSet(TRIGGER_PARAM, value)}
      onDateChange={(value) => writeParam(SINCE_PARAM, value)}
      onClearFilters={handleClearFilters}
    />
  )
}
