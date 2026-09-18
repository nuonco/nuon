import { useState } from 'react'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit/BranchRunCommit'
import { Button } from '@/components/common/Button'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'

export type TActivityFilter = 'all' | 'awaiting-approval' | 'failed' | 'in-progress'

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

function matchesFilter(item: TBranchActivityItem, filter: TActivityFilter): boolean {
  if (filter === 'all') return true
  const statuses = ATTENTION_STATUSES[filter]
  return statuses.includes(item.runStatus)
}

export const BranchActivityFeed = ({ items, isLoading = false }: IBranchActivityFeed) => {
  const [activeFilter, setActiveFilter] = useState<TActivityFilter>('all')

  const filtered = items.filter((item) => matchesFilter(item, activeFilter))

  const emptyMessage =
    activeFilter === 'all'
      ? 'Branch activity will appear here once app branches start processing runs.'
      : `No branches with ${FILTERS.find((f) => f.value === activeFilter)?.label.toLowerCase()} status.`

  return (
    <div className="flex flex-col gap-4">
      <div className="flex items-center gap-1" role="group" aria-label="Filter branch activity">
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
        <TimelineSkeleton eventCount={5} />
      ) : filtered.length === 0 ? (
        <EmptyState
          variant="history"
          size="sm"
          emptyTitle={activeFilter === 'all' ? 'No branch activity yet' : 'No matching branches'}
          emptyMessage={emptyMessage}
        />
      ) : (
        <Timeline
          events={filtered.map((item) => ({ ...item, created_at: item.runCreatedAt }))}
          pagination={{ limit: filtered.length, offset: 0, hasNext: false }}
          getEventKey={(item) => item.runId}
          renderEvent={(item) => (
            <TimelineEvent
              key={item.runId}
              status={item.runStatus}
              createdAt={item.runCreatedAt}
              title={
                <span className="flex items-center gap-1.5 min-w-0">
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
              }
              underline={
                <BranchRunCommit
                  status={item.runStatus}
                  href={item.runHref ?? item.commitHref}
                  message={item.commitMessage}
                  sha={item.commitSha}
                  author={item.commitAuthor}
                  avatarUrl={item.commitAvatarUrl}
                  createdAt={item.runCreatedAt}
                />
              }
            />
          )}
        />
      )}
    </div>
  )
}
