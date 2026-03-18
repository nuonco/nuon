import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import type { TPaginationMeta } from '@/lib/api'

export interface IActivity {
  id: string
  installName: string
  installId: string
  message: string
  status: string
  created_at: string
  duration?: string
  triggeredBy: string
  href?: string
}

export const RecentActivities = ({
  activities,
  pagination,
}: {
  activities: IActivity[]
  pagination?: TPaginationMeta
}) => {
  if (!activities || activities.length === 0) {
    return (
      <EmptyState
        variant="history"
        emptyTitle="No recent activity"
        emptyMessage="Activity will appear here once your runner starts processing jobs."
        size="sm"
      />
    )
  }

  return (
    <Timeline
      events={activities}
      pagination={{
        limit: pagination?.limit ?? 10,
        offset: pagination?.offset ?? 0,
        hasNext: pagination?.hasNext ?? false,
      }}
      renderEvent={(activity) => (
        <TimelineEvent
          key={activity.id}
          status={activity.status}
          createdAt={activity.created_at}
          createdBy={activity.triggeredBy || undefined}
          title={
            <span className="flex items-center gap-2">
              {activity.href ? (
                <Link
                  href={activity.href}
                  className="text-primary-600 dark:text-primary-400"
                >
                  {activity.installName}
                </Link>
              ) : (
                <Text variant="body" weight="strong">
                  {activity.installName}
                </Text>
              )}
              <Text variant="body" theme="neutral">
                {activity.message}
              </Text>
            </span>
          }
          caption={
            activity.duration ? `Completed in ${activity.duration}` : undefined
          }
        />
      )}
    />
  )
}
