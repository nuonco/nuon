import { Badge } from '@/components/common/Badge'
import { Duration } from '@/components/common/Duration'
import { EmptyState } from '@/components/common/EmptyState'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { Timeline } from '@/components/common/Timeline'
import { TimelineEvent } from '@/components/common/TimelineEvent'
import { TimelineSkeleton } from '@/components/common/TimelineSkeleton'
import { Tooltip } from '@/components/common/Tooltip'
import { BranchRunCommit } from '@/components/branches/BranchRunCommit'
import { RunDeploymentGraph } from '@/components/branches/RunDeploymentGraph'
import {
  getRunTitle,
  getRunTrigger,
} from '@/components/branches/shared/run-title'
import {
  getBranchRunFromWorkflow,
  isPreviewWorkflow,
} from '@/components/branches/shared/preview-run-utils'
import type { TInstall, TInstallGroupRun, TWorkflow } from '@/types'
import {
  getWorkflowBadge,
  getWorkflowPendingApprovals,
  isBranchRunWorkflow,
  isServiceAccount,
} from '@/utils/workflow-utils'
import { useWorkflowApprovals } from '@/hooks/use-workflow-approvals'
import { CancelWorkflowButton } from '../CancelWorkflow'

export interface IWorkflowTimeline {
  workflows: TWorkflow[]
  pagination: { hasNext: boolean; offset: number; limit: number }
  orgId: string
  installId?: string
  install?: TInstall
  branchRunGraphs?: Record<
    string,
    {
      installGroupRuns: TInstallGroupRun[]
      installsById?: Record<string, TInstall>
    }
  >
  isLoading?: boolean
  isFiltered?: boolean
  getWorkflowHref?: (workflow: TWorkflow) => string
}

export const WorkflowTimeline = ({
  workflows,
  pagination,
  orgId,
  installId,
  install,
  branchRunGraphs,
  isLoading,
  isFiltered = false,
  getWorkflowHref,
}: IWorkflowTimeline) => {
  const { approvals } = useWorkflowApprovals()

  if (isLoading) return <TimelineSkeleton eventCount={10} />

  return workflows?.length ? (
    <Timeline<TWorkflow>
      events={workflows}
      pagination={pagination}
      renderEvent={(workflow) => {
        const isBranchRun = isBranchRunWorkflow(workflow)
        const branchRun = getBranchRunFromWorkflow(workflow)
        const trigger = getRunTrigger(branchRun)
        const commit = branchRun?.vcs_connection_commit
        const runGraph = workflow.id
          ? branchRunGraphs?.[workflow.id]
          : undefined
        const workflowHref = getWorkflowHref
          ? getWorkflowHref(workflow)
          : `/${orgId}/installs/${installId}/workflows/${workflow.id}`
        const createdByAccount = workflow?.created_by
        const createdBy = createdByAccount?.email ? (
          isServiceAccount(createdByAccount) ? (
            <Tooltip
              position="left"
              tipContent={
                <Text variant="subtext" family="mono">
                  {createdByAccount.email}
                </Text>
              }
            >
              a service account
            </Tooltip>
          ) : (
            createdByAccount.email
          )
        ) : undefined

        const workflowTitle = (
          <span className="flex items-center gap-4 mb-1">
            <Link
              variant="inline"
              className="inline-flex gap-2 items-center"
              href={workflowHref}
            >
              {isBranchRun
                ? getRunTitle(workflow)
                : workflow.name || workflow?.type || workflow.id}
            </Link>
            {workflow?.status?.status === 'in-progress' ? (
              <Badge size="sm" theme="info">
                In progress
              </Badge>
            ) : null}
            {workflow?.approval_option === 'prompt' &&
            workflow?.status?.status !== 'approval-awaiting' &&
            getWorkflowPendingApprovals(approvals, workflow?.id).length ? (
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
              <span className="flex items-center gap-2 flex-wrap">
                {isBranchRun && isPreviewWorkflow(workflow) ? (
                  <>
                    <Badge variant="code" size="sm">
                      preview
                    </Badge>
                    {trigger === 'manual' ? (
                      <Badge variant="code" size="sm">
                        manual
                      </Badge>
                    ) : branchRun?.preview?.source === 'commit' ? (
                      <Badge variant="code" size="sm">
                        commit
                      </Badge>
                    ) : null}
                  </>
                ) : workflow.plan_only ? (
                  isBranchRun ? (
                    <Badge variant="code" size="sm">
                      preview
                    </Badge>
                  ) : (
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
                  )
                ) : null}
                {isBranchRun &&
                !isPreviewWorkflow(workflow) &&
                trigger === 'manual' ? (
                  <Badge variant="code" size="sm">
                    manual
                  </Badge>
                ) : null}
                {isBranchRun &&
                !isPreviewWorkflow(workflow) &&
                trigger === 'github_label' &&
                branchRun?.metadata?.github_label ? (
                  <LabelBadge
                    labelKey="label"
                    labelValue={branchRun.metadata.github_label}
                    size="sm"
                  />
                ) : null}
                {workflow?.type === 'drift_run_reprovision_sandbox' ||
                workflow.type === 'drift_run' ? (
                  <Badge variant="code" size="sm">
                    cron scheduled
                  </Badge>
                ) : null}
                {workflow?.approval_option === 'approve-all' &&
                workflow?.metadata?.approval_type ? (
                  <Badge variant="code" size="sm">
                    {workflow.metadata.approval_type === 'approve-workflow'
                      ? 'auto-approve (workflow)'
                      : workflow.metadata.approval_type === 'install-config'
                        ? 'auto-approve (config)'
                        : 'auto-approve'}
                  </Badge>
                ) : null}
              </span>
            }
            badge={getWorkflowBadge(workflow)}
            caption={isBranchRun ? undefined : <ID>{workflow?.id}</ID>}
            underline={
              isBranchRun ? (
                <span className="flex flex-col gap-4 mt-2">
                  {commit ? (
                    <BranchRunCommit
                      status={branchRun?.status ?? workflow?.status?.status}
                      href={workflowHref}
                      message={commit.message?.split('\n')[0]}
                      author={commit.author_name}
                      avatarUrl={commit.author_avatar_url}
                      sha={commit.sha}
                      createdAt={commit.created_at ?? workflow.created_at}
                      className="ml-2 border-l pl-4 py-1"
                      showStatus={false}
                    />
                  ) : branchRun?.head_sha ? (
                    <Text
                      family="mono"
                      variant="subtext"
                      theme="neutral"
                      className="ml-2 border-l pl-4 py-1"
                    >
                      {branchRun.head_sha.slice(0, 7)}
                    </Text>
                  ) : null}
                  {runGraph ? (
                    <span className="pl-6 pt-1">
                      <RunDeploymentGraph
                        installGroupRuns={runGraph.installGroupRuns}
                        installsById={runGraph.installsById}
                        orgId={orgId}
                      />
                    </span>
                  ) : null}
                </span>
              ) : (
                <span className="flex items-center gap-6 mt-1">
                  <Text
                    flex
                    className="gap-1"
                    variant="subtext"
                    theme="neutral"
                    title="Created"
                  >
                    <Icon variant="CalendarBlankIcon" />{' '}
                    <Time time={workflow?.created_at} variant="subtext" />
                  </Text>
                  <Text
                    flex
                    className="gap-1"
                    variant="subtext"
                    theme="neutral"
                    title={workflow?.finished ? 'Finished' : 'Last updated'}
                  >
                    <Icon variant="ClockClockwiseIcon" />{' '}
                    <Time
                      time={
                        workflow?.finished && workflow?.finished_at
                          ? workflow.finished_at
                          : workflow?.updated_at
                      }
                      variant="subtext"
                      format="relative"
                    />
                  </Text>
                  {workflow?.finished && workflow?.execution_time ? (
                    <Text
                      flex
                      className="gap-1"
                      variant="subtext"
                      theme="neutral"
                      title="Duration"
                    >
                      <Icon variant="TimerIcon" />{' '}
                      <Duration
                        nanoseconds={workflow?.execution_time}
                        variant="subtext"
                      />
                    </Text>
                  ) : null}
                </span>
              )
            }
            createdAt={workflow?.created_at}
            createdBy={
              isBranchRun && trigger !== 'manual' ? undefined : createdBy
            }
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
        emptyMessage={
          isFiltered
            ? 'Clear a filter to widen the results.'
            : 'Workflow activity will appear here when a run starts.'
        }
        emptyTitle={
          isFiltered
            ? 'No runs match these filters'
            : 'No workflow activity yet'
        }
      />
    </div>
  )
}
