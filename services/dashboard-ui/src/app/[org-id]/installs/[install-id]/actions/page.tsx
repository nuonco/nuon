import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  InstallPageSubNav,
  InstallStatuses,
  InstallActionWorkflowsTable,
  DashboardContent,
  ErrorFallback,
  Loading,
  NoActions,
  Section,  
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { getInstall, getInstallActionWorkflowLatestRun } from '@/lib'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Actions`,
  }
}

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install = await getInstall({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        { href: `/${orgId}/installs/${install.id}`, text: install.name },
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
      meta={
        <InstallPageSubNav
          installId={installId}
          orgId={orgId}
          runnerId={install?.runner_id}
        />
      }
    >
      <Section childrenClassName="flex flex-auto">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading actions..." />
            }
          >
            <LoadInstallActions installId={installId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
})

const LoadInstallActions: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const actionsWithLatestRun = await getInstallActionWorkflowLatestRun({
    installId,
    orgId,
  })

  return actionsWithLatestRun?.length ? (
    <InstallActionWorkflowsTable
      actions={actionsWithLatestRun}
      installId={installId}
      orgId={orgId}
    />
  ) : (
    <NoActions />
  )
}
