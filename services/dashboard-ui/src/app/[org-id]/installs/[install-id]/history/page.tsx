import { type FC, Suspense } from 'react'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  Loading,
  InstallHistory,
  InstallManagementDropdown,
  InstallPageSubNav,
  InstallStatuses,
  Section,
  Text,
  Time,
} from '@/components'
import {
  getInstall,
  getInstallEvents,
  getOrg,
  getAppLatestInputConfig,
} from '@/lib'
import { USER_REPROVISION, INSTALL_UPDATE } from '@/utils'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getOrg({ orgId }),
  ])

  const appInputConfigs =
    (await getAppLatestInputConfig({
      appId: install?.app_id,
      orgId,
    }).catch(console.error)) || undefined

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}`,
          text: install.name,
        },
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
              installId={installId}
              orgId={orgId}
              hasInstallComponents={Boolean(
                install?.install_components?.length
              )}
              hasUpdateInstall={INSTALL_UPDATE}
              inputConfig={appInputConfigs}
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
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section heading="History" className="overflow-auto">
          <Suspense
            fallback={
              <Loading
                loadingText="Loading install history..."
                variant="page"
              />
            }
          >
            <LoadInstallHistory installId={installId} orgId={orgId} />
          </Suspense>
        </Section>
      </div>
    </DashboardContent>
  )
})

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
