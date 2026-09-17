import type { TWorkflow } from '@/types/ctl-api.types'
import { Text } from '../../atoms/Text'
import { FilterDropdown } from '../../molecules/FilterMenu'
import { ListSearch } from '../../molecules/ListSearch'
import { Pagination } from '../../molecules/Pagination'
import { WorkflowTimelineItem } from '../../molecules/WorkflowTimelineItem'
import { Timeline } from '../Timeline/Timeline'

export interface IWorkflowFilter<T extends string> {
  label: string
  options: readonly { value: T; label: string }[]
  selected: Set<T>
  onToggle: (value: T) => void
  onIsolate: (value: T) => void
  onReset: () => void
}

export interface IWorkflowTimeline {
  workflows: TWorkflow[]
  search: string
  onSearchChange: (value: string) => void
  offset: number
  pageSize: number
  hasNext: boolean
  onOffsetChange: (offset: number) => void
  statusFilter: IWorkflowFilter<string>
  typeFilter: IWorkflowFilter<string>
  previewFilter?: IWorkflowFilter<string>
  dateFilter: IWorkflowFilter<string>
  getWorkflowHref?: (workflow: TWorkflow) => string | undefined
  pendingApprovalIds?: ReadonlySet<string>
  driftedWorkflowIds?: ReadonlySet<string>
  filtered?: boolean
  loading?: boolean
  fetching?: boolean
  error?: unknown
}

const EmptyState = ({
  filtered,
  error,
}: {
  filtered: boolean
  error: unknown
}) => {
  if (error) {
    return (
      <span className="flex flex-col items-center gap-1">
        <Text weight="medium">Activity failed to load</Text>
        <Text variant="caption" color="tertiary">
          Retry in a moment, or reload the page.
        </Text>
      </span>
    )
  }

  if (filtered) {
    return (
      <span className="flex flex-col items-center gap-1">
        <Text weight="medium">No runs match these filters</Text>
        <Text variant="caption" color="tertiary">
          Clear a filter to widen the results.
        </Text>
      </span>
    )
  }

  return (
    <span className="flex flex-col items-center gap-1">
      <Text weight="medium">No activity yet</Text>
      <Text variant="caption" color="tertiary">
        Runs appear here as soon as something changes on this branch.
      </Text>
    </span>
  )
}

const filterControl = (filter: IWorkflowFilter<string>) => (
  <FilterDropdown
    key={filter.label}
    label={filter.label}
    options={filter.options.map((option) => ({
      value: option.value,
      label: option.label,
      textValue: option.label,
    }))}
    selected={filter.selected}
    onToggle={filter.onToggle}
    onIsolate={filter.onIsolate}
    onReset={filter.onReset}
    constrained={filter.selected.size > 0}
  />
)

export const WorkflowTimeline = ({
  workflows,
  search,
  onSearchChange,
  offset,
  pageSize,
  hasNext,
  onOffsetChange,
  statusFilter,
  typeFilter,
  previewFilter,
  dateFilter,
  getWorkflowHref,
  pendingApprovalIds,
  driftedWorkflowIds,
  filtered = false,
  loading = false,
  fetching = false,
  error,
}: IWorkflowTimeline) => (
  <div className="flex min-w-0 flex-col gap-4">
    <div className="flex min-w-0 flex-wrap items-center gap-2">
      <ListSearch
        value={search}
        onValueChange={onSearchChange}
        placeholder="Search by title or ID"
        aria-label="Search activity"
        className="w-full max-w-sm"
      />
      {filterControl(statusFilter)}
      {filterControl(typeFilter)}
      {previewFilter ? filterControl(previewFilter) : null}
      {filterControl(dateFilter)}
    </div>

    <Timeline
      events={workflows}
      getEventKey={(workflow, index) => workflow?.id ?? index}
      getEventTime={(workflow) => workflow?.created_at}
      loading={loading}
      loadingLabel="Loading activity"
      emptyState={<EmptyState filtered={filtered} error={error} />}
    >
      {(workflow) => (
        <WorkflowTimelineItem
          workflow={workflow}
          href={getWorkflowHref?.(workflow)}
          pendingApprovalIds={pendingApprovalIds}
          driftedWorkflowIds={driftedWorkflowIds}
        />
      )}
    </Timeline>

    <Pagination
      label="Activity pagination"
      offset={offset}
      pageSize={pageSize}
      hasNext={hasNext}
      loading={fetching}
      onOffsetChange={onOffsetChange}
    />
  </div>
)
