import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Divider } from '@/components/common/Divider'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { getInstallWorkflows } from '@/lib'
import type { TWorkflow } from '@/types'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getWorkflowBadge } from '@/utils/workflow-utils'
import { CancelWorkflowButton } from './CancelWorkflow'

const POLL_INTERVAL = 20000

export const ActiveWorkflows = ({ installId }: { installId: string }) => {
  const { org } = useOrg()
  const { install } = useInstall()

  const { data } = useQuery({
    queryKey: ['install-active-workflows', org?.id, installId],
    queryFn: () =>
      getInstallWorkflows({
        orgId: org.id,
        installId,
        finished: false,
        limit: 50,
        offset: 0,
      }),
    refetchInterval: POLL_INTERVAL,
    enabled: !!org?.id && !!installId,
  })

  const workflows = (data?.data ?? []).filter(
    (w) =>
      w.status?.status &&
      w.status.status !== 'pending' &&
      w.status.status !== 'queued'
  )

  if (!workflows.length) return null

  return (
    <div className="flex flex-col gap-6">
      <Text variant="base" weight="strong">
        Active workflows
      </Text>
      <Timeline<TWorkflow>
        events={workflows}
        groupByDate={false}
        pagination={{ hasNext: false, offset: 0, limit: 50 }}
        renderEvent={(workflow) => {
          const workflowTitle = (
            <Link
              className="inline-flex gap-2 items-center"
              href={`/${org.id}/installs/${installId}/workflows/${workflow.id}`}
            >
              {workflow?.type === 'action_workflow_run' &&
              workflow?.metadata?.adhoc_action
                ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
                : workflow.name || toSentenceCase(snakeToWords(workflow.type))}
            </Link>
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
                <span className="flex items-center gap-6">
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
      <Divider />
    </div>
  )
}
