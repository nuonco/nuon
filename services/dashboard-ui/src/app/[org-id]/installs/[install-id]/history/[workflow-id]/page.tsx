import type { Metadata } from 'next'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
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
  const installId = params?.['install-id'] as string
  const installWorkflowId = params?.['workflow-id'] as string
  const orgId = params?.['org-id'] as string
  const [install, installWorkflow] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallWorkflow({ installWorkflowId, orgId }),
  ])

  return {
    title: `${install.name} | ${
      installWorkflow?.name ||
      removeSnakeCase(sentanceCase(installWorkflow?.type))
    }`,
  }
}

export default withPageAuthRequired(async function InstallWorkflow({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const installWorkflowId = params?.['workflow-id'] as string
  const [install, installWorkflow] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallWorkflow({ installWorkflowId, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
        {
          href: `/${orgId}/installs/${install.id}/history`,
          text: 'History',
        },
        {
          href: `/${orgId}/installs/${install.id}/history/${installWorkflowId}`,
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
        </span>
      }
      headingUnderline={install?.id}
      meta={
        <div className="pb-6 flex flex-col gap-4">
          <div className="flex gap-8">
            <div className="flex flex-col gap-1">
              <Text variant="reg-12" isMuted>
                Pending approvals
              </Text>
              <Text variant="med-18">
                {
                  installWorkflow?.steps?.filter(
                    (s) =>
                      s?.execution_type === 'approval' && !s?.approval?.response
                  )?.length
                }
              </Text>
            </div>

            <div className="flex flex-col gap-1">
              <Text variant="reg-12" isMuted>
                Total steps
              </Text>
              <Text variant="med-18">{installWorkflow?.steps?.length}</Text>
            </div>

            <div className="flex flex-col gap-1">
              <Text variant="reg-12" isMuted>
                Completed steps
              </Text>
              <Text variant="med-18">
                {
                  installWorkflow?.steps?.filter(
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
          {installWorkflow?.steps?.some(
            (s) => s?.execution_type === 'approval'
          ) ? (
            <div className="flex flex-col gap-3">
              {installWorkflow?.approval_option === 'prompt' &&
              !installWorkflow?.finished ? (
                <>
                  <Text>
                    Automatically approve all changes waiting for approval
                  </Text>
                  <WorkflowApproveAllModal workflow={installWorkflow} />
                </>
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
          {!installWorkflow?.finished && installWorkflow?.steps?.length > 0 ? (
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
})
