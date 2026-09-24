import { useState } from 'react'
import {
  BranchPlanDots,
  type TBranchPlanGroup,
} from '@/components/branches/BranchCards/BranchPlanDots'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit/BranchRunCommit'
import { MiniDeploymentView } from '@/components/branches/MiniDeploymentView'
import { Button } from '@/components/common/Button'
import { Card } from '@/components/common/Card'
import { EmptyState } from '@/components/common/EmptyState'
import { Expand } from '@/components/common/Expand'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Skeleton } from '@/components/common/Skeleton'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'

export type TActivityFilter =
  | 'all'
  | 'awaiting-approval'
  | 'failed'
  | 'in-progress'

export interface TRunUpdatedInstall {
  id: string
  name: string
  group: string
  href?: string
  runStatus?: string
  health?: string
  rolledOut?: boolean
}

export interface TBranchActivityItem {
  appId: string
  appName: string
  appHref?: string
  branchId: string
  branchName: string
  branchHref?: string
  runId: string
  runStatus: string
  runCreatedAt: string
  runHref?: string
  commitMessage?: string
  commitSha?: string
  commitAuthor?: string
  commitAvatarUrl?: string
  commitHref?: string
  planGroups?: TBranchPlanGroup[]
  updatedInstalls?: TRunUpdatedInstall[]
}

export interface IBranchActivityFeed {
  items: TBranchActivityItem[]
  isLoading?: boolean
}

const FILTERS: { label: string; value: TActivityFilter }[] = [
  { label: 'All', value: 'all' },
  { label: 'Awaiting approval', value: 'awaiting-approval' },
  { label: 'Failed', value: 'failed' },
  { label: 'In progress', value: 'in-progress' },
]

const ATTENTION_STATUSES: Record<Exclude<TActivityFilter, 'all'>, string[]> = {
  'awaiting-approval': ['awaiting-approval', 'pending-approval'],
  failed: ['failed', 'error'],
  'in-progress': ['in-progress', 'running', 'queued'],
}

function matchesFilter(
  item: TBranchActivityItem,
  filter: TActivityFilter
): boolean {
  if (filter === 'all') return true
  const statuses = ATTENTION_STATUSES[filter]
  return statuses.includes(item.runStatus)
}

const UpdatedInstalls = ({
  installs,
  runId,
  planGroups,
}: {
  installs: TRunUpdatedInstall[]
  runId: string
  planGroups: TBranchPlanGroup[]
}) => (
  <Expand
    id={`${runId}-updated-installs`}
    heading={<BranchPlanDots groups={planGroups} />}
    isIconBeforeHeading
    headerClassName="px-1 py-1 rounded"
  >
    <div className="pl-7 pr-1 pb-2 pt-1">
      <MiniDeploymentView groups={planGroups} installs={installs} />
    </div>
  </Expand>
)

const RunPlan = ({ item }: { item: TBranchActivityItem }) => {
  const planGroups = item.planGroups ?? []
  if (planGroups.length === 0) return null

  const installs = item.updatedInstalls ?? []
  if (installs.length === 0) {
    return (
      <div className="flex items-center gap-2 px-1 py-1">
        <span aria-hidden className="w-4 shrink-0" />
        <BranchPlanDots groups={planGroups} />
      </div>
    )
  }

  return (
    <UpdatedInstalls
      installs={installs}
      runId={item.runId}
      planGroups={planGroups}
    />
  )
}

const UpdateCard = ({ item }: { item: TBranchActivityItem }) => (
  <Card className="gap-3 p-4 min-w-0" data-run-id={item.runId}>
    <div className="flex items-start justify-between gap-4">
      <div className="flex items-center gap-2 min-w-0">
        <Status
          status={item.runStatus}
          variant="timeline"
          isWithoutText
          className="shrink-0"
        />
        <span className="flex items-center gap-1.5 min-w-0 flex-wrap">
          {item.appHref ? (
            <Link href={item.appHref} variant="inline">
              {item.appName}
            </Link>
          ) : (
            <Text variant="body" weight="strong" as="span">
              {item.appName}
            </Text>
          )}
          <Text variant="subtext" theme="neutral" as="span">
            /
          </Text>
          <Icon variant="GitBranchIcon" size={13} theme="neutral" />
          {item.branchHref ? (
            <Link href={item.branchHref} variant="inline">
              {item.branchName}
            </Link>
          ) : (
            <Text variant="body" weight="strong" as="span">
              {item.branchName}
            </Text>
          )}
        </span>
      </div>
      <div className="flex items-center gap-2 shrink-0">
        {item.runHref ? (
          <Button href={item.runHref} variant="secondary" size="sm">
            View run
          </Button>
        ) : null}
        <Time
          time={item.runCreatedAt}
          format="relative"
          variant="subtext"
          theme="neutral"
          className="shrink-0"
        />
      </div>
    </div>

    <BranchRunCommit
      status={item.runStatus}
      href={item.runHref ?? item.commitHref}
      message={item.commitMessage}
      sha={item.commitSha}
      author={item.commitAuthor}
      avatarUrl={item.commitAvatarUrl}
      createdAt={item.runCreatedAt}
      showStatus={false}
    />

    <RunPlan item={item} />
  </Card>
)

export const BranchActivityFeed = ({
  items,
  isLoading = false,
}: IBranchActivityFeed) => {
  const [activeFilter, setActiveFilter] = useState<TActivityFilter>('all')

  const filtered = items.filter((item) => matchesFilter(item, activeFilter))

  const emptyMessage =
    activeFilter === 'all'
      ? 'Branch activity will appear here once app branches start processing runs.'
      : `No branches with ${FILTERS.find((f) => f.value === activeFilter)?.label.toLowerCase()} status.`

  return (
    <div className="flex flex-col gap-4">
      <div
        className="flex items-center gap-1"
        role="group"
        aria-label="Filter branch activity"
      >
        {FILTERS.map(({ label, value }) => (
          <Button
            key={value}
            variant="ghost"
            size="sm"
            isActive={activeFilter === value}
            onClick={() => setActiveFilter(value)}
            aria-pressed={activeFilter === value}
          >
            {label}
          </Button>
        ))}
      </div>

      {isLoading ? (
        <div className="flex flex-col gap-3">
          {Array.from({ length: 3 }).map((_, index) => (
            <Card key={index} className="gap-3 p-4">
              <Skeleton lines={3} width={['35%', '70%', '100%']} />
            </Card>
          ))}
        </div>
      ) : filtered.length === 0 ? (
        <EmptyState
          variant="history"
          size="sm"
          emptyTitle={
            activeFilter === 'all'
              ? 'No branch activity yet'
              : 'No matching branches'
          }
          emptyMessage={emptyMessage}
        />
      ) : (
        <div className="flex flex-col gap-3">
          {filtered.map((item) => (
            <UpdateCard key={item.runId} item={item} />
          ))}
        </div>
      )}
    </div>
  )
}
