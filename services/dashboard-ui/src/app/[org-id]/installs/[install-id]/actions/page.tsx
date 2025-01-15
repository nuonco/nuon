// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  InstallPageSubNav,
  InstallStatuses,
  InstallActionWorkflowsTable,
  DashboardContent,
  Section,
} from '@/components'
import { getInstall, getInstallActionWorkflowLatestRun, getOrg } from '@/lib'

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

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        { href: `/${org.id}/installs/${install.id}`, text: install.name },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
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
