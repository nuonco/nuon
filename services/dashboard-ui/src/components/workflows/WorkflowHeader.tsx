'use client'

import { BackLink } from '@/components/common/BackLink'
import { Badge } from '@/components/old/Badge'
import { Link } from '@/components/old/Link'
import { Notice } from '@/components/old/Notice'
import { Text, ID } from '@/components/old/Typography'
import { YAStatus } from '@/components/old/InstallWorkflows/InstallWorkflowHistory'
import { WorkflowApproveAllModal } from '@/components/old/InstallWorkflows/ApproveAllModal'
import { InstallWorkflowCancelModal } from '@/components/old/InstallWorkflows/InstallWorkflowCancelModal'
import { InstallWorkflowActivity } from '@/components/old/InstallWorkflows/InstallWorkflowActivity'
import { useInstall } from '@/hooks/use-install'
import { useOrg } from '@/hooks/use-org'
import { usePolling, type IPollingProps } from '@/hooks/use-polling'
import type { TWorkflow } from '@/types'
import { removeSnakeCase, sentanceCase } from '@/utils'

interface IWorkflowHeader extends IPollingProps {
  initWorkflow: TWorkflow
}

export const WorkflowHeader = ({
  initWorkflow,
  pollInterval = 5000,
  shouldPoll = false,
}: IWorkflowHeader) => {
  const { org } = useOrg()
  const { install } = useInstall()
  const { data: workflow, error } = usePolling<TWorkflow>({
    initData: initWorkflow,
    path: `/api/orgs/${org.id}/workflows/${initWorkflow?.id}`,
    pollInterval,
    shouldPoll,
  })

  const workflowSteps =
    workflow?.steps?.filter((s) => s?.execution_type !== 'hidden') || []

  return (
    <header className="px-6 py-8 flex flex-col pt-6 gap-6 border-b">
      <BackLink />

      {error?.error ? (
        <Notice>{error?.error || 'Unable to load workflow'}</Notice>
      ) : null}
      <div className="flex items-start justify-between gap-4">
        <hgroup className="flex flex-col gap-2">
          <Text level={1} role="heading" variant="semi-18">
            <span className="flex gap-2 items-center">
              <YAStatus status={workflow?.status?.status} />
              {workflow?.name || removeSnakeCase(sentanceCase(workflow?.type))}
              {workflow?.type === 'action_workflow_run' &&
              workflow?.metadata?.install_action_workflow_name
                ? ` (${workflow?.metadata?.install_action_workflow_name})`
                : ' '}
              {workflow?.plan_only ? (
                <Badge
                  className="!text-[11px] my-2"
                  variant="code"
                  theme="info"
                >
                  Drift scan
                </Badge>
              ) : null}

              {workflow?.type === 'drift_run_reprovision_sandbox' ||
              workflow.type === 'drift_run' ? (
                <Badge className="!text-[11px]" variant="code">
                  cron scheduled
                </Badge>
              ) : null}
            </span>
          </Text>
          <span>
            <span className="flex gap-4">
              <ID id={workflow?.id} />
              <Text variant="reg-12">
                <Link href={`/${org.id}/apps/${install.app_id}`}>
                  {install?.app?.name}
                </Link>
              </Text>
            </span>
            <div className="flex flex-col gap-2 mt-4">
              <div className="flex gap-8">
                <div className="flex flex-col gap-1">
                  <Text variant="reg-12" isMuted>
                    Pending approvals
                  </Text>
                  <Text variant="med-18" isMuted>
                    {
                      workflowSteps.filter(
                        (s) =>
                          s?.execution_type === 'approval' &&
                          !s?.approval?.response &&
                          s?.status?.status !== 'discarded'
                      )?.length
                    }
                  </Text>
                </div>

                <div className="flex flex-col gap-1">
                  <Text variant="reg-12" isMuted>
                    Total steps
                  </Text>
                  <Text variant="med-18" isMuted>
                    {workflowSteps.length}
                  </Text>
                </div>

                <div className="flex flex-col gap-1">
                  <Text variant="reg-12" isMuted>
                    Completed steps
                  </Text>
                  <Text variant="med-18" isMuted>
                    {
                      workflowSteps.filter(
                        (s) =>
                          s?.status?.status === 'success' ||
                          s?.status?.status === 'active' ||
                          s?.status?.status === 'error' ||
                          s?.status?.status === 'approved'
                      )?.length
                    }
                  </Text>
                </div>
              </div>
              {workflowSteps.some(
                (s) => s?.execution_type === 'approval' && !workflow?.plan_only
              ) ? (
                <div className="flex flex-col gap-4">
                  <div className="flex flex-col gap-3">
                    {workflow?.approval_option === 'prompt' &&
                    !workflow?.finished ? (
                      <></>
                    ) : workflowSteps.some(
                        (s) => s?.approval?.response?.type === 'deny'
                      ) ? (
                      <Text className="text-red-600 dark:text-red-400">
                        Changes have been denied
                      </Text>
                    ) : (
                      <Text className="text-green-600 dark:text-green-400">
                        All changes have been approved
                      </Text>
                    )}
                  </div>
                </div>
              ) : null}
            </div>
          </span>
        </hgroup>

        <div className="flex flex-col gap-3 items-end">
          <div className="flex items-center gap-4">
            {workflow?.steps?.length > 2 &&
            workflow?.approval_option === 'prompt' &&
            !workflow?.finished &&
            workflow?.status?.status !== 'cancelled' &&
            !workflow?.plan_only ? (
              <WorkflowApproveAllModal
                workflow={workflow}
                buttonVariant="primary"
              />
            ) : null}
            {!workflow?.finished && workflow?.status?.status !== 'cancelled' ? (
              <InstallWorkflowCancelModal installWorkflow={workflow} />
            ) : null}
          </div>

          <InstallWorkflowActivity installWorkflow={workflow} />
        </div>
      </div>
    </header>
  )
}
