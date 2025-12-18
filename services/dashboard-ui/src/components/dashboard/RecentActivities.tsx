'use client'

import { ArrowRight } from '@phosphor-icons/react'
import { Button } from '@/components/common/Button'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'

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
  orgId,
}: {
  activities: IActivity[]
  orgId: string
}) => {
  if (!activities || activities.length === 0) {
    return (
      <div className="py-8 text-center">
        <Text theme="neutral">No recent activities</Text>
      </div>
    )
  }

  return (
    <div className="flex flex-col">
      <Timeline
        events={activities}
        pagination={{ limit: 10, offset: 0, hasNext: false }}
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
            caption={activity.duration ? `Completed in ${activity.duration}` : undefined}
          />
        )}
      />
      <div className="pt-4 flex justify-center">
        <Button variant="ghost" size="sm" href={`/${orgId}/installs`}>
          View all
          <ArrowRight size={14} weight="bold" />
        </Button>
      </div>
    </div>
  )
}
