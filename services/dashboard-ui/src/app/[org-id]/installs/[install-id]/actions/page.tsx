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
  Section,
  Text,
  Time,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import { getInstall, getInstallActionWorkflowLatestRun } from '@/lib'
import { USER_REPROVISION } from '@/utils'

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
      statues={
        <div className="flex items-start gap-8">
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Created
            </Text>
            <Time variant="reg-12" time={install?.created_at} />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Updated
            </Text>
            <Time variant="reg-12" time={install?.updated_at} />
          </span>
          <InstallStatuses initInstall={install} shouldPoll />
          {USER_REPROVISION ? (
            <InstallManagementDropdown
              orgId={orgId}
              hasInstallComponents={Boolean(
                install?.install_components?.length
              )}
              install={install}
            />
          ) : null}
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
      <Section heading="All workflows">
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
    <Text variant="reg-14">No actions configured on this app.</Text>
  )
}
