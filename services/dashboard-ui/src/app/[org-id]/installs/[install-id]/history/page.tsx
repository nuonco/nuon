import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  Loading,
  InstallHistory,
  InstallPageSubNav,
  InstallStatuses,
  InstallWorkflowHistory,
  Section,
  Text,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import {
  getInstall,
  getInstallEvents,
  getInstallWorkflows,
  getOrg,
} from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | History`,
  }
}

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      headingMeta={
        <>
          Last updated <Time time={install?.updated_at} format="relative" />
        </>
      }
      statues={
        <div className="flex items-start gap-8">
          <InstallStatuses initInstall={install} shouldPoll />

          <InstallManagementDropdown
            orgId={orgId}
            hasInstallComponents={Boolean(install?.install_components?.length)}
            install={install}
          />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="Install history" className="overflow-auto">
          <Suspense
            fallback={
              <Loading
                loadingText="Loading install history..."
                variant="page"
              />
            }
          >
            {org?.features?.['install-independent-runner'] ? (
              <LoadInstallWorkflows installId={installId} orgId={orgId} />
            ) : (
              <LoadInstallHistory installId={installId} orgId={orgId} />
            )}
          </Suspense>
        </Section>
      </div>
    </DashboardContent>
  )
})

const LoadInstallWorkflows: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const installWorkflows = await getInstallWorkflows({
    installId,
    orgId,
  }).catch(console.error)

  return installWorkflows ? (
    <InstallWorkflowHistory installWorkflows={installWorkflows} shouldPoll />
  ) : (
    <Text>No install history yet.</Text>
  )
}

const LoadInstallHistory: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const events = await getInstallEvents({ installId, orgId })
  return (
    <InstallHistory
      initEvents={events}
      installId={installId}
      orgId={orgId}
      shouldPoll
    />
  )
}
