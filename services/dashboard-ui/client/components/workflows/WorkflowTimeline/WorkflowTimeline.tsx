import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import type { TInstall, TWorkflow } from '@/types'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import {
  getWorkflowBadge,
  getPendingApprovalCount,
} from '@/utils/workflow-utils'
import { CancelWorkflowButton } from '../CancelWorkflow'

export interface IWorkflowTimeline {
  workflows: TWorkflow[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  installId: string
  install?: TInstall
}

export const WorkflowTimeline = ({
  workflows,
  pagination,
  orgId,
  installId,
  install,
}: IWorkflowTimeline) => {
  return workflows?.length ? (
    <Timeline<TWorkflow>
      events={workflows}
      pagination={pagination}
      renderEvent={(workflow) => {
        const workflowTitle = (
          <span className="flex items-center gap-4 mb-1">
            <Link
              className="inline-flex gap-2 items-center"
              href={`/${orgId}/installs/${installId}/workflows/${workflow.id}`}
            >
              {workflow?.type === 'action_workflow_run' &&
              workflow?.metadata?.adhoc_action
                ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
                : workflow.name || toSentenceCase(snakeToWords(workflow.type))}
            </Link>
            {workflow?.status?.status === 'in-progress' ? (
              <Badge size="sm" theme="info">
                In progress
              </Badge>
            ) : null}
            {workflow?.approval_option === 'prompt' &&
            getPendingApprovalCount(workflow) ? (
              <Badge size="sm" theme="warn">
                Pending approval
              </Badge>
            ) : null}
          </span>
        )

        return (
          <TimelineEvent
            key={workflow.id}
            actions={
              !workflow?.finished &&
              workflow?.status?.status !== 'cancelled' &&
              workflow?.status?.status !== 'error' ? (
                <CancelWorkflowButton workflow={workflow} size="sm" />
              ) : null
            }
            additionalCaption={
              <span className="flex items-center gap-2">
                {workflow.plan_only ? (
                  <>
                    <Badge variant="code" size="sm">
                      drift scan
                    </Badge>
                    {install?.drifted_objects &&
                    install?.drifted_objects?.find(
                      (d) => d?.install_workflow_id === workflow?.id
                    ) ? (
                      <Badge size="sm" variant="code" theme="warn">
                        drift detected
                      </Badge>
                    ) : null}
                  </>
                ) : null}
                {workflow?.type === 'drift_run_reprovision_sandbox' ||
                workflow.type === 'drift_run' ? (
                  <Badge variant="code" size="sm">
                    cron scheduled
                  </Badge>
                ) : null}
              </span>
            }
            badge={getWorkflowBadge(workflow)}
            caption={<ID>{workflow?.id}</ID>}
            underline={
              <span className="flex items-center gap-6 mt-1">
                <Text
                  flex
                  className="gap-1"
                  variant="subtext"
                  theme="neutral"
                >
                  <Icon variant="CalendarBlankIcon" />{' '}
                  <Time time={workflow?.created_at} variant="subtext" />
                </Text>
                <Text
                  flex
                  className="gap-1"
                  variant="subtext"
                  theme="neutral"
                >
                  <Icon variant="ClockClockwiseIcon" />{' '}
                  <Time
                    time={workflow?.updated_at}
                    variant="subtext"
                    format="relative"
                  />
                </Text>
                {workflow?.finished ? (
                  <Text
                    flex
                    className="gap-1"
                    variant="subtext"
                    theme="neutral"
                  >
                    <Icon variant="TimerIcon" />{' '}
                    <Duration
                      nanoseconds={workflow?.execution_time}
                      variant="subtext"
                    />
                  </Text>
                ) : null}
              </span>
            }
            createdAt={workflow?.created_at}
            createdBy={workflow?.created_by?.email}
            status={workflow?.status?.status}
            title={workflowTitle}
          />
        )
      }}
    />
  ) : (
    <div className="mx-auto mt-24">
      <EmptyState
        variant="table"
        emptyMessage="There are no workflows to display. This could be because no workflows have run yet, or your current filters are not matching any results."
        emptyTitle="No workflows found"
      />
    </div>
  )
}

export const WorkflowTimelineSkeleton = () => {
  return <TimelineSkeleton eventCount={10} />
}
