import { Badge } from '@/components/common/Badge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import type { TInstallScheduledReconcile } from '@/types'
import {
  SCHEDULED_RECONCILE_STATUS_LABEL,
  SCHEDULED_RECONCILE_STATUS_THEME,
} from './schedule-utils'

export interface IScheduledReconcileHistory {
  history: TInstallScheduledReconcile[]
  getWorkflowHref?: (workflowId: string) => string | undefined
}

export const ScheduledReconcileHistory = ({
  history,
  getWorkflowHref,
}: IScheduledReconcileHistory) => {
  if (history.length === 0) return null

  return (
    <div className="flex flex-col gap-2">
      <Text variant="subtext" theme="neutral" weight="strong">
        Recent schedules
      </Text>
      <div className="flex flex-col gap-2">
        {history.map((item) => {
          const status = item.status ?? 'cancelled'
          const workflowHref =
            item.workflow_id && getWorkflowHref
              ? getWorkflowHref(item.workflow_id)
              : undefined
          return (
            <div
              key={item.id}
              className="flex items-center justify-between gap-3 py-1.5 border-b last:border-b-0 !border-cool-grey-200 dark:!border-dark-grey-700"
            >
              <div className="flex flex-col gap-0.5 min-w-0">
                <Text variant="subtext">
                  <Time time={item.scheduled_at} format="short-datetime" />
                </Text>
                {item.status_reason ? (
                  <Text variant="subtext" theme="neutral" className="truncate">
                    {item.status_reason}
                  </Text>
                ) : null}
              </div>
              <div className="flex items-center gap-2 shrink-0">
                {item.plan_only ? (
                  <Badge size="sm" theme="neutral">
                    Plan only
                  </Badge>
                ) : null}
                {workflowHref ? (
                  <Link href={workflowHref}>
                    <Badge
                      size="sm"
                      theme={SCHEDULED_RECONCILE_STATUS_THEME[status]}
                    >
                      {SCHEDULED_RECONCILE_STATUS_LABEL[status]}
                    </Badge>
                  </Link>
                ) : (
                  <Badge
                    size="sm"
                    theme={SCHEDULED_RECONCILE_STATUS_THEME[status]}
                  >
                    {SCHEDULED_RECONCILE_STATUS_LABEL[status]}
                  </Badge>
                )}
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
