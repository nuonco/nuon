'use client'

import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { Link } from '@/components/common/Link'
import { Timeline, type ITimeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import { useQueryParams } from '@/hooks/use-query-params'
import type { TWorkflow } from '@/types'
import { toSentenceCase, snakeToWords } from '@/utils/string-utils'
import { getWorkflowBadge } from '@/utils/workflow-utils'
import { CancelWorkflowButton } from './CancelWorkflow'

interface IWorkflowTimeline
  extends Omit<ITimeline<TWorkflow>, 'events' | 'renderEvent'>,
    IPollingProps {
  initWorkflows: Array<TWorkflow>
  ownerId: string
  ownerType: 'apps' | 'installs'
  type?: string
  planonly?: boolean
}

export const WorkflowTimeline = ({
  initWorkflows,
  pagination,
  shouldPoll = false,
  pollInterval = 20000,
  ownerId,
  ownerType,
  planonly = true,
  type = '',
}: IWorkflowTimeline) => {
  const { org } = useOrg()
  const queryParams = useQueryParams({
    offset: pagination?.offset,
    limit: 10,
    planonly,
    type,
  })
  const { data: workflows } = usePolling<TWorkflow[]>({
    dependencies: [queryParams],
    path: `/api/orgs/${org?.id}/${ownerType}/${ownerId}/workflows${queryParams}`,
    shouldPoll,
    initData: initWorkflows,
    pollInterval,
  })

  return workflows?.length ? (
    <Timeline<TWorkflow>
      events={workflows}
      pagination={pagination}
      renderEvent={(workflow) => {
        const workflowTitle = (
          <Link
            className="inline-flex gap-2 items-center"
            href={`/${org.id}/${ownerType}/${ownerId}/workflows/${workflow.id}`}
          >
            {workflow.name || toSentenceCase(snakeToWords(workflow.type))}
          </Link>
        )

        return (
          <TimelineEvent
            key={workflow.id}
            actions={
              !workflow?.finished &&
              workflow?.status?.status !== 'cancelled' ? (
                <CancelWorkflowButton workflow={workflow} size="sm" />
              ) : null
            }
            additionalCaption={
              <span className="flex items-center gap-2">
                {workflow.plan_only ? (
                  <Badge variant="code" size="sm">
                    drift scan
                  </Badge>
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
            caption={workflow?.id}
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
        emptyMessage="There are no workflows to display. This could be because no workflows have run yet, 
        or your current filters are not matching any results."
        emptyTitle="No workflows founds"
      />
    </div>
  )
}

export const WorkflowTimelineSkeleton = () => {
  return <TimelineSkeleton eventCount={10} />
}
