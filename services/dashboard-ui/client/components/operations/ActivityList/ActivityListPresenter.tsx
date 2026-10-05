import { useEffect } from 'react'
import { Badge } from '@/components/common/Badge'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { CheckboxFilterDropdown } from '@/components/common/CheckboxFilterDropdown'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon, type TIconVariant } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Pagination, type IPagination } from '@/components/common/Pagination'
import { RadioFilterDropdown } from '@/components/common/RadioFilterDropdown'
import { SearchInput } from '@/components/common/SearchInput'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { WorkflowPanelLink } from '@/components/workflows/InstallWorkflowPanel'
import { useInstallLink } from '@/hooks/use-install-path'
import { usePagination } from '@/hooks/use-pagination'
import { useSurfaces } from '@/hooks/use-surfaces'
import { PaginationProvider } from '@/providers/pagination-provider'
import type { TAPIError, TInstallActivity, TInstallActivityType } from '@/types'
import {
  WORKFLOW_DATE_LABELS,
  WORKFLOW_STATUS_LABELS,
  type TWorkflowDatePreset,
  type TWorkflowStatusOption,
} from '@/utils/workflow-filters'
import { ActivityDetailPanel } from './ActivityDetailPanel'

export const ACTIVITY_TYPE_LABELS: Record<TInstallActivityType, string> = {
  action_run: 'Action run',
  runbook_run: 'Runbook run',
  policy_check: 'Policy check',
}

export const ACTIVITY_STATUS_OPTIONS = [
  'running',
  'queued',
  'succeeded',
  'failed',
  'cancelled',
  'warning',
] as const satisfies readonly TWorkflowStatusOption[]

const ACTIVITY_TYPE_ICON: Record<TInstallActivityType, TIconVariant> = {
  action_run: 'TerminalWindowIcon',
  runbook_run: 'BookIcon',
  policy_check: 'ShieldCheckIcon',
}

const OFFSET_PARAM = 'activity_offset'

interface IActivityCard {
  activity: TInstallActivity
  orgId: string
  installId: string
  onViewDetails: () => void
}

const ActivityCard = ({
  activity,
  orgId,
  installId,
  onViewDetails,
}: IActivityCard) => {
  const installLink = useInstallLink()
  const subjectHref = activity.action?.name
    ? installLink({
        orgId,
        installId,
        suffix: `/operations/actions?q=${encodeURIComponent(activity.action.name)}`,
      })
    : activity.runbook?.name
      ? installLink({
          orgId,
          installId,
          suffix: `/operations/runbooks?q=${encodeURIComponent(activity.runbook.name)}`,
        })
      : undefined
  const subjectLabel =
    activity.type === 'runbook_run'
      ? 'Runbook'
      : activity.type === 'policy_check'
        ? 'Policy'
        : 'Action'
  const subjectName =
    activity.action?.name ||
    activity.runbook?.name ||
    activity.policy?.component_name

  return (
    <Card className="!p-4 !gap-3 !shadow-none">
      <div className="flex items-start justify-between gap-4">
        <div className="flex items-start gap-3 min-w-0">
          <span className="mt-0.5 text-cool-grey-400 shrink-0">
            <Icon variant={ACTIVITY_TYPE_ICON[activity.type]} size={16} />
          </span>
          <div className="flex flex-col gap-1 min-w-0">
            <div className="flex items-center gap-2 flex-wrap">
              <Text weight="strong">{activity.title}</Text>
              <Status status={activity.status} variant="badge" />
              <Badge size="sm" theme="neutral">
                {ACTIVITY_TYPE_LABELS[activity.type]}
              </Badge>
            </div>
            {activity.summary && (
              <Text variant="subtext" theme="neutral">
                {activity.summary}
              </Text>
            )}
          </div>
        </div>
        <div className="flex items-center gap-3 shrink-0">
          <Time
            time={activity.created_at}
            format="relative"
            variant="subtext"
            theme="neutral"
          />
          <Button variant="secondary" size="sm" onClick={onViewDetails}>
            View details
          </Button>
        </div>
      </div>

      {(subjectHref || activity.workflow || activity.policy) && (
        <div className="flex items-center gap-x-6 gap-y-2 flex-wrap">
          {subjectHref && subjectName && (
            <span className="flex items-center gap-2">
              <Text as="span" variant="subtext" theme="neutral">
                {subjectLabel}
              </Text>
              <Link href={subjectHref} textVariant="subtext">
                {subjectName}
              </Link>
            </span>
          )}
          {activity.workflow && (
            <span className="flex items-center gap-2">
              <Text as="span" variant="subtext" theme="neutral">
                Workflow
              </Text>
              <WorkflowPanelLink
                workflowId={activity.workflow.id}
                textVariant="subtext"
              >
                {activity.workflow.name}
              </WorkflowPanelLink>
            </span>
          )}
          {activity.policy && (
            <Text variant="subtext" theme="neutral">
              {activity.policy.deny_count} denied, {activity.policy.warn_count}{' '}
              warnings, {activity.policy.pass_count} passed
            </Text>
          )}
        </div>
      )}
    </Card>
  )
}

const ActivityCardSkeleton = () => (
  <Card className="!p-4 !gap-3 !shadow-none">
    <div className="flex items-start justify-between gap-4">
      <div className="flex items-start gap-3 min-w-0">
        <Text variant="subtext" loading loadingWidth={2} className="mt-0.5" />
        <div className="flex flex-col gap-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <Text loading loadingWidth={22} />
            <Status loading variant="badge" loadingWidth={8} />
            <Badge loading size="sm" loadingWidth={12} />
          </div>
          <Text variant="subtext" loading loadingWidth={44} />
        </div>
      </div>
      <div className="flex items-center gap-3 shrink-0">
        <Text variant="subtext" loading loadingWidth={10} />
        <Badge loading size="lg" loadingWidth={11} className="!rounded-lg" />
      </div>
    </div>
    <div className="flex items-center gap-x-6 gap-y-2 flex-wrap">
      <Text variant="subtext" loading loadingWidth={16} />
      <Text variant="subtext" loading loadingWidth={20} />
    </div>
  </Card>
)

export interface IActivityFilter {
  search: string
  status: Set<TWorkflowStatusOption>
  type: Set<TInstallActivityType>
  date?: TWorkflowDatePreset
}

export interface IActivityListPresenter {
  activity: TInstallActivity[]
  isLoading: boolean
  error?: TAPIError | null
  pagination: Omit<IPagination, 'position'>
  orgId: string
  installId: string
  search: string
  filter: IActivityFilter
  onSearchChange: (value: string) => void
  onStatusChange: (value: Set<TWorkflowStatusOption>) => void
  onTypeChange: (value: Set<TInstallActivityType>) => void
  onDateChange: (value?: string) => void
  onClearFilters: () => void
}

const ActivityListBase = ({
  activity,
  isLoading,
  error,
  pagination,
  orgId,
  installId,
  search,
  filter,
  onSearchChange,
  onStatusChange,
  onTypeChange,
  onDateChange,
  onClearFilters,
}: IActivityListPresenter) => {
  const { addPanel } = useSurfaces()
  const { isPaginating, setIsPaginating } = usePagination()

  useEffect(() => {
    setIsPaginating(false)
  }, [activity, setIsPaginating])

  const hasActiveFilters =
    filter.search !== '' ||
    filter.status.size > 0 ||
    filter.type.size > 0 ||
    !!filter.date

  return (
    <div className="flex flex-col gap-4">
      <div className="flex min-w-0 flex-wrap items-center gap-2">
        <SearchInput
          aria-label="Search activity"
          placeholder="Search activity"
          value={search}
          onChange={onSearchChange}
          onClear={() => onSearchChange('')}
        />
        <CheckboxFilterDropdown
          id="activity-filter-status"
          label="Status"
          options={ACTIVITY_STATUS_OPTIONS.map((value) => ({
            value,
            label: WORKFLOW_STATUS_LABELS[value],
          }))}
          selected={filter.status}
          onChange={(value) =>
            onStatusChange(value as Set<TWorkflowStatusOption>)
          }
        />
        <CheckboxFilterDropdown
          id="activity-filter-type"
          label="Type"
          options={(
            Object.keys(ACTIVITY_TYPE_LABELS) as TInstallActivityType[]
          ).map((value) => ({
            value,
            label: ACTIVITY_TYPE_LABELS[value],
          }))}
          selected={filter.type}
          onChange={(value) => onTypeChange(value as Set<TInstallActivityType>)}
        />
        <RadioFilterDropdown
          id="activity-filter-date"
          label="Date"
          options={(
            Object.keys(WORKFLOW_DATE_LABELS) as TWorkflowDatePreset[]
          ).map((value) => ({
            value,
            label: WORKFLOW_DATE_LABELS[value],
          }))}
          selected={filter.date}
          onChange={onDateChange}
        />
        {hasActiveFilters && (
          <Button variant="ghost" onClick={onClearFilters}>
            Clear filters
          </Button>
        )}
      </div>

      {isPaginating || (isLoading && activity.length === 0) ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: pagination.limit ?? 5 }).map((_, index) => (
            <ActivityCardSkeleton key={index} />
          ))}
        </div>
      ) : error ? (
        <EmptyState
          emptyTitle="Activity failed to load"
          emptyMessage="Unable to load activity. Try refreshing the page."
          className="my-12"
        />
      ) : activity.length === 0 ? (
        <EmptyState
          emptyTitle={
            hasActiveFilters ? 'No activity found' : 'No activity yet'
          }
          emptyMessage={
            hasActiveFilters
              ? 'No activity matches the current filters. Try adjusting or clearing them.'
              : 'Activity will appear here once an action, runbook, or policy check runs on this install.'
          }
          className="my-12"
        />
      ) : (
        <div className="flex flex-col gap-3">
          {activity.map((item) => (
            <ActivityCard
              key={item.id}
              activity={item}
              orgId={orgId}
              installId={installId}
              onViewDetails={() => {
                addPanel(
                  <ActivityDetailPanel
                    activity={item}
                    orgId={orgId}
                    installId={installId}
                  />
                )
              }}
            />
          ))}
        </div>
      )}

      {pagination.hasNext || pagination.offset !== 0 ? (
        <Pagination {...pagination} param={OFFSET_PARAM} />
      ) : null}
    </div>
  )
}

export const ActivityListPresenter = (props: IActivityListPresenter) => (
  <PaginationProvider>
    <ActivityListBase {...props} />
  </PaginationProvider>
)
