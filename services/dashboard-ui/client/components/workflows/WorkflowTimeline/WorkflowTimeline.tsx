import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
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
          <div className="flex items-center gap-2 flex-wrap">
            <Link
              href={`/${orgId}/installs/${installId}/workflows/${workflow.id}`}
              className="font-medium"
            >
              {workflow?.type === 'action_workflow_run' &&
              workflow?.metadata?.adhoc_action
                ? `Adhoc action run (${workflow?.metadata?.install_action_workflow_name})`
                : workflow.name || toSentenceCase(snakeToWords(workflow.type))}
            </Link>
            {workflow?.status?.status === 'in-progress' ? (
              <Badge theme="info">
                In progress
              </Badge>
            ) : null}
            {workflow?.approval_option === 'prompt' &&
            getPendingApprovalCount(workflow) ? (
              <Badge theme="warn">
                Pending approval
              </Badge>
            ) : null}
          </div>
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
                    <Badge variant="code">
                      drift scan
                    </Badge>
                    {install?.drifted_objects &&
                    install?.drifted_objects?.find(
                      (d) => d?.install_workflow_id === workflow?.id
                    ) ? (
                      <Badge variant="code" theme="warn">
                        drift detected
                      </Badge>
                    ) : null}
                  </>
                ) : null}
                {workflow?.type === 'drift_run_reprovision_sandbox' ||
                workflow.type === 'drift_run' ? (
                  <Badge variant="code">
                    cron scheduled
                  </Badge>
                ) : null}
              </span>
            }
            badge={getWorkflowBadge(workflow)}
            caption={<ID>{workflow?.id}</ID>}
            createdAt={workflow?.created_at}
            createdBy={workflow?.created_by?.email}
            duration={workflow?.finished ? workflow?.execution_time : undefined}
            status={workflow?.status?.status}
            title={workflowTitle}
            updatedAt={workflow?.updated_at}
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
