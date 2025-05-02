import type { Metadata } from 'next'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  Empty,
  InstallWorkflowActivity,
  InstallWorkflowSteps,
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
    title: `${install.name} | ${installWorkflow?.name}`,
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
      heading={installWorkflow?.name}
      headingUnderline={install?.id}
      statues={
        <InstallWorkflowActivity
          installWorkflow={installWorkflow}
          shouldPoll
          pollDuration={3000}
        />
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
