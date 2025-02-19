// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  InstallPageSubNav,
  InstallStatuses,
  InstallActionWorkflowsTable,
  InstallManagementDropdown,
  DashboardContent,
  Section,
  Text,
  Time,
} from '@/components'
import {
  getInstall,
  getInstallActionWorkflowLatestRun,
  getOrg,
  getAppLatestInputConfig,
} from '@/lib'
import { USER_REPROVISION, INSTALL_UPDATE } from '@/utils'

export default withPageAuthRequired(async function InstallWorkflowRuns({
  params,
}) {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const [org, install, actionsWithLatestRun] = await Promise.all([
    getOrg({ orgId }),
    getInstall({ installId, orgId }),
    getInstallActionWorkflowLatestRun({ installId, orgId }).catch(
      console.error
    ),
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
        { href: `/${org.id}/installs/${install.id}`, text: install.name },
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
      <Section className="border-r" heading="All workflows">
        {actionsWithLatestRun?.length ? (
          <InstallActionWorkflowsTable
            actions={actionsWithLatestRun}
            installId={installId}
            orgId={orgId}
          />
        ) : (
          'No actions configured on this app'
        )}
      </Section>
    </DashboardContent>
  )
})
