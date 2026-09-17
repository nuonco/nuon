import { useMemo } from 'react'
import { useSSETimelineQuery } from '@/lib/sse/use-sse-timeline-query'
import { getBranchWorkflowRuns, getInstallWorkflows } from '@/lib'
import type { TWorkflow } from '@/types/ctl-api.types'
import { useListQueryState } from '../../../hooks/use-list-query-state'
import { useToast } from '../../../hooks/use-toast'
import { commaSetQueryParameter } from '../../../utils/list-query'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_PREVIEW_LABELS,
  WORKFLOW_STATUS_LABELS,
  WORKFLOW_TYPE_LABELS,
  datePresetFilterParameter,
  previewFilterParameter,
  statusFilterParameter,
  typeFilterParameter,
  workflowStatusOptions,
  workflowTypeOptions,
  type TWorkflowDatePreset,
  type TWorkflowPreviewOption,
  type TWorkflowStatusOption,
  type TWorkflowTypeGroup,
} from '../../../utils/workflow-filters'
import { WorkflowTimeline, type IWorkflowFilter } from './WorkflowTimeline'

const PAGE_SIZE = 20
const POLL_INTERVAL = 20_000

const ACTIVITY_FILTERS = {
  status: commaSetQueryParameter<TWorkflowStatusOption>('status'),
  type: commaSetQueryParameter<TWorkflowTypeGroup>('type'),
  date: commaSetQueryParameter<TWorkflowDatePreset>('since'),
  preview: commaSetQueryParameter<TWorkflowPreviewOption>('preview'),
}

export type TWorkflowTimelineOwner =
  | { kind: 'app'; appId: string; branchId: string }
  | { kind: 'install'; installId: string }

export interface IWorkflowTimelineContainer {
  orgId: string | undefined
  owner: TWorkflowTimelineOwner
  pendingApprovalIds?: ReadonlySet<string>
  driftedWorkflowIds?: ReadonlySet<string>
}

const ownerReady = (owner: TWorkflowTimelineOwner) =>
  owner.kind === 'app' ? !!owner.appId && !!owner.branchId : !!owner.installId

const ownerQueryKey = (owner: TWorkflowTimelineOwner) =>
  owner.kind === 'app'
    ? ['branch-runs', owner.appId, owner.branchId]
    : ['install-workflows', owner.installId]

const ownerSSEPath = (orgId: string, owner: TWorkflowTimelineOwner) =>
  owner.kind === 'app'
    ? `/api/orgs/${orgId}/apps/${owner.appId}/branches/${owner.branchId}/runs/sse`
    : `/api/orgs/${orgId}/installs/${owner.installId}/workflows/sse`

const ownerEventName = (owner: TWorkflowTimelineOwner) =>
  owner.kind === 'app' ? 'branch-runs' : 'workflows'

const filterControl = <T extends string>({
  label,
  options,
  labels,
  selected,
  onChange,
}: {
  label: string
  options: readonly T[]
  labels: Record<T, string>
  selected: Set<T>
  onChange: (value: Set<T>) => void
}): IWorkflowFilter<string> => ({
  label,
  options: options.map((value) => ({ value, label: labels[value] })),
  selected: selected as Set<string>,
  onToggle: (value) => {
    const next = new Set(selected)
    if (next.has(value as T)) next.delete(value as T)
    else next.add(value as T)
    onChange(next)
  },
  onIsolate: (value) =>
    onChange(
      selected.size === 1 && selected.has(value as T)
        ? new Set<T>()
        : new Set<T>([value as T])
    ),
  onReset: () => onChange(new Set<T>()),
})

export const WorkflowTimelineContainer = ({
  orgId,
  owner,
  pendingApprovalIds,
  driftedWorkflowIds,
}: IWorkflowTimelineContainer) => {
  const { addToast } = useToast()
  const list = useListQueryState({
    pageSize: PAGE_SIZE,
    filters: ACTIVITY_FILTERS,
  })

  const enabled = !!orgId && ownerReady(owner)
  const status = statusFilterParameter(list.filters.status)
  const type = typeFilterParameter(list.filters.type)
  const search = list.search || undefined
  const datePreset = [...list.filters.date].at(0)
  const preview =
    owner.kind === 'app'
      ? previewFilterParameter(list.filters.preview)
      : undefined

  const sseUrl = useMemo(() => {
    if (!orgId || !ownerReady(owner)) return undefined

    const params = new URLSearchParams({
      limit: String(list.pageSize),
      offset: String(list.offset),
    })
    if (search) params.set(owner.kind === 'app' ? 'q' : 'search', search)
    if (status) params.set('status', status)
    if (type) params.set('type', type)
    const createdAtGte = datePresetFilterParameter(list.filters.date)
    if (createdAtGte) params.set('created_at_gte', createdAtGte)
    if (preview !== undefined) params.set('preview', String(preview))

    return `${ownerSSEPath(orgId, owner)}?${params}`
  }, [
    datePreset,
    preview,
    list.offset,
    list.pageSize,
    orgId,
    owner,
    search,
    status,
    type,
  ])

  const fetchWorkflows = () => {
    const createdAtGte = datePresetFilterParameter(list.filters.date)

    if (owner.kind === 'app') {
      return getBranchWorkflowRuns({
        orgId: orgId!,
        appId: owner.appId,
        branchId: owner.branchId,
        limit: list.pageSize,
        offset: list.offset,
        q: search,
        status,
        type,
        preview,
        created_at_gte: createdAtGte,
      })
    }

    return getInstallWorkflows({
      orgId: orgId!,
      installId: owner.installId,
      limit: list.pageSize,
      offset: list.offset,
      search,
      status,
      type,
      created_at_gte: createdAtGte,
    })
  }

  const { data, isLoading, error } = useSSETimelineQuery({
    sseUrl,
    queryKey: ['activity', orgId, ...ownerQueryKey(owner), ...list.queryKey],
    queryFn: fetchWorkflows,
    enabled,
    shouldPoll: true,
    pollInterval: POLL_INTERVAL,
    eventName: ownerEventName(owner),
    onError: (message: string) =>
      addToast({
        heading: 'Refresh failed',
        theme: 'warn',
        description: message,
      }),
  })

  const result = data as
    | { data?: TWorkflow[]; pagination?: { hasNext?: boolean } }
    | undefined

  const filtered =
    !!search ||
    list.filters.status.size > 0 ||
    list.filters.type.size > 0 ||
    list.filters.date.size > 0 ||
    preview !== undefined

  return (
    <WorkflowTimeline
      workflows={result?.data ?? []}
      search={list.search}
      onSearchChange={list.setSearch}
      offset={list.offset}
      pageSize={list.pageSize}
      hasNext={result?.pagination?.hasNext ?? false}
      onOffsetChange={list.setOffset}
      statusFilter={filterControl({
        label: 'Status',
        options: workflowStatusOptions(),
        labels: WORKFLOW_STATUS_LABELS,
        selected: list.filters.status,
        onChange: (value) => list.setFilter('status', value),
      })}
      typeFilter={filterControl({
        label: 'Type',
        options: workflowTypeOptions(owner.kind),
        labels: WORKFLOW_TYPE_LABELS,
        selected: list.filters.type,
        onChange: (value) => list.setFilter('type', value),
      })}
      previewFilter={
        owner.kind === 'app'
          ? filterControl({
              label: 'Preview',
              options: ['preview', 'rollout'] as const,
              labels: WORKFLOW_PREVIEW_LABELS,
              selected: list.filters.preview,
              onChange: (value) => list.setFilter('preview', value),
            })
          : undefined
      }
      dateFilter={filterControl({
        label: 'Date',
        options: ['24h', '7d', '30d'] as const,
        labels: WORKFLOW_DATE_LABELS,
        selected: list.filters.date,
        onChange: (value) => list.setFilter('date', value),
      })}
      pendingApprovalIds={pendingApprovalIds}
      driftedWorkflowIds={driftedWorkflowIds}
      filtered={filtered}
      loading={isLoading}
      error={error}
    />
  )
}
