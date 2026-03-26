import { Badge } from '@/components/common/Badge'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { useOrg } from '@/hooks/use-org'
import type { TInstall, TWorkflow } from '@/types'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import {
  getWorkflowBadge,
  getPendingApprovalCount,
} from '@/utils/workflow-utils'
import { CancelWorkflowButton } from './CancelWorkflow'

export const ActiveWorkflows = ({
  workflows,
  install,
  hasDivider = false,
}: {
  workflows: TWorkflow[]
  install?: TInstall
  hasDivider?: boolean
}) => {
  const { org } = useOrg()

  const inProgressWorkflows = workflows.filter(
    (w) => w?.status?.status === 'in-progress'
  )

  if (!inProgressWorkflows.length) return null

  return (
    <div className="flex flex-col gap-6">
      <Text variant="base" weight="strong">
        Active workflows
      </Text>
      <Timeline<TWorkflow>
        events={inProgressWorkflows}
        groupByDate={false}
        pagination={{ hasNext: false, offset: 0, limit: 50 }}
        renderEvent={(workflow) => {
          const installId = workflow.owner_id
          const installName = workflow.metadata?.owner_name
          const workflowTitle = (
            <span className="flex items-center gap-4">
              <Link
                className="inline-flex gap-2 items-center"
                href={`/${org.id}/installs/${installId}/workflows/${workflow.id}`}
              >
                {installName && !install && (
                  <>
                    <Text>{installName}</Text>
                    <span>|</span>
                  </>
                )}
                {workflow?.type === 'action_workflow_run' &&
                workflow?.metadata?.adhoc_action
                  ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
                  : workflow.name ||
                    toSentenceCase(snakeToWords(workflow.type))}
              </Link>
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
              caption={
                <span className="flex items-center gap-6 mt-2">
                  <ID>{workflow?.id}</ID>
                  <Text
                    className="!flex gap-1"
                    variant="subtext"
                    theme="neutral"
                  >
                    <Icon variant="CalendarBlankIcon" />{' '}
                    <Time time={workflow?.created_at} variant="subtext" />
                  </Text>
                  <Text
                    className="!flex gap-1"
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
      {hasDivider ? <Divider /> : null}
    </div>
  )
}
