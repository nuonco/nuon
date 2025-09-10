import type { Metadata } from 'next'
import {
  Badge,
  DashboardContent,
  Empty,
  InstallWorkflowActivity,
  InstallWorkflowSteps,
  InstallWorkflowCancelModal,
  WorkflowApproveAllModal,
  YAStatus,
  Text,
} from '@/components'
import { getInstall, getInstallWorkflow } from '@/lib'

import { removeSnakeCase, sentanceCase } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['workflow-id']: installWorkflowId,
  } = await params
  const [install, installWorkflow] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallWorkflow({ installWorkflowId, orgId }),
  ])

  return {
    title: `${install?.name} | ${
      installWorkflow?.name ||
      removeSnakeCase(sentanceCase(installWorkflow?.type))
    }`,
  }
}

export default async function InstallWorkflow({ params }) {
  const {
    ['org-id']: orgId,
    ['install-id']: installId,
    ['workflow-id']: installWorkflowId,
  } = await params
  const [install, installWorkflow] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallWorkflow({ installWorkflowId, orgId }),
  ])

  const workflowSteps =
    installWorkflow?.steps?.filter((s) => s?.execution_type !== 'hidden') || []

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/workflows`,
          text: 'Workflows',
        },
        {
          href: `/${orgId}/installs/${install.id}/workflows/${installWorkflowId}`,
          text:
            installWorkflow?.name ||
            removeSnakeCase(sentanceCase(installWorkflow?.type)),
        },
      ]}
      heading={
        <span className="flex gap-2 items-center">
          <YAStatus status={installWorkflow?.status?.status} />
          {installWorkflow?.name ||
            removeSnakeCase(sentanceCase(installWorkflow?.type))}
          {installWorkflow?.type === 'action_workflow_run' &&
          installWorkflow?.metadata?.install_action_workflow_name
            ? ` (${installWorkflow?.metadata?.install_action_workflow_name})`
            : ' '}
          {installWorkflow?.plan_only ? (
            <Badge className="!text-[11px] ml-2" variant="code">
              Plan only
            </Badge>
          ) : null}
        </span>
      }
      headingUnderline={installWorkflow?.id}
      meta={
        <div className="pb-6 flex flex-col gap-4">
          <div className="flex gap-8">
            <div className="flex flex-col gap-1">
              <Text variant="reg-12" isMuted>
                Pending approvals
              </Text>
              <Text variant="med-18">
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
              <Text variant="med-18">{workflowSteps.length}</Text>
            </div>

            <div className="flex flex-col gap-1">
              <Text variant="reg-12" isMuted>
                Completed steps
              </Text>
              <Text variant="med-18">
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
          {workflowSteps.some((s) => s?.execution_type === 'approval') ? (
            <div className="flex flex-col gap-3">
              {installWorkflow?.approval_option === 'prompt' &&
              !installWorkflow?.finished ? (
                <>
                  <Text>
                    Automatically approve all changes waiting for approval
                  </Text>
                  <WorkflowApproveAllModal workflow={installWorkflow} />
                </>
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
          ) : null}
        </div>
      }
      statues={
        <div className="flex flex-col gap-3 items-end">
          {!installWorkflow?.finished &&
          installWorkflow?.status?.status !== 'cancelled' ? (
            <InstallWorkflowCancelModal installWorkflow={installWorkflow} />
          ) : null}

          <InstallWorkflowActivity
            installWorkflow={installWorkflow}
            shouldPoll
            pollDuration={3000}
          />
        </div>
      }
    >
      <>
        {installWorkflow ? (
          <InstallWorkflowSteps
            installWorkflow={installWorkflow}
            orgId={orgId}
            install={install}
          />
        ) : (
          <Empty
            emptyTitle="No install history"
            emptyMessage="Waiting on this install to start provisioning"
            variant="history"
          />
        )}
      </>
    </DashboardContent>
  )
}
