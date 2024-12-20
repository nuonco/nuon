// TODO(nnnat): remove once we have this API change on prod
// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  InstallHistory,
  InstallPageSubNav,
  InstallStatuses,
  Section,
  InstallReprovisionButton,
} from '@/components'
import {
  getInstall,
  getInstallEvents,
  getInstallRunnerGroup,
  getOrg,
} from '@/lib'
import { USER_REPROVISION } from '@/utils'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, events, runnerGroup, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallEvents({ installId, orgId }),
    getInstallRunnerGroup({ installId, orgId }),
    getOrg({ orgId }),
  ])

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
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <Section
          heading="History"
          className="overflow-auto history"
          actions={
            USER_REPROVISION ? (
              <InstallReprovisionButton installId={installId} orgId={orgId} />
            ) : null
          }
        >
          <InstallHistory
            initEvents={events}
            installId={installId}
            orgId={orgId}
            shouldPoll
          />
        </Section>
      </div>
    </DashboardContent>
  )
})
