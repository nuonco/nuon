import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  Empty,
  Loading,
  InstallPageSubNav,
  InstallStatuses,
  InstallWorkflowActivity,
  InstallWorkflowSteps,
  InstallManagementDropdown,
  Section,
  Time,
} from '@/components'
import { getInstall, getInstallWorkflow } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Workflow`,
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
          text: installWorkflow?.name,
        },
      ]}
      heading={installWorkflow?.name}
      headingUnderline={installWorkflow.id}
    >
      <div className="grid grid-cols-1 md:grid-cols-12 divide-x flex-auto">
        <div className="flex flex-col lg:flex-row flex-auto col-span-8">
          <Section heading="Install update steps" className="overflow-auto">
            <Suspense
              fallback={
                <Loading
                  loadingText="Loading install workflow..."
                  variant="page"
                />
              }
            >
              <LoadInstallWorkflow
                installId={installId}
                installWorkflowId={installWorkflowId}
                orgId={orgId}
              />
            </Suspense>
          </Section>
        </div>
        <div className="col-span-4">
          <Section className="overflow-auto">
            <Suspense
              fallback={
                <Loading
                  loadingText="Loading install activity..."
                  variant="stack"
                />
              }
            >
              <LoadInstallWorkflowActivity
                installId={installId}
                installWorkflowId={installWorkflowId}
                orgId={orgId}
              />
            </Suspense>
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})

const LoadInstallWorkflow: FC<{
  installId: string
  installWorkflowId: string
  orgId: string
}> = async ({ installWorkflowId, orgId }) => {
  const installWorkflow = await getInstallWorkflow({
    installWorkflowId,
    orgId,
  }).catch(console.error)

  return installWorkflow ? (
    <InstallWorkflowSteps installWorkflow={installWorkflow} orgId={orgId} />
  ) : (
    <Empty
      emptyTitle="No install history"
      emptyMessage="Waiting on this install to start provisioning"
      variant="history"
    />
  )
}

const LoadInstallWorkflowActivity: FC<{
  installId: string
  installWorkflowId: string
  orgId: string
}> = async ({ installWorkflowId, orgId }) => {
  const installWorkflow = await getInstallWorkflow({
    installWorkflowId,
    orgId,
  }).catch(console.error)

  return installWorkflow ? (
    <InstallWorkflowActivity installWorkflow={installWorkflow} shouldPoll />
  ) : (
    <div className="p-4 border rounded-md">
      <Empty
        emptyTitle="No install activity"
        emptyMessage="Waiting on this install to start provisioning"
        variant="history"
        isSmall
      />
    </div>
  )
}
