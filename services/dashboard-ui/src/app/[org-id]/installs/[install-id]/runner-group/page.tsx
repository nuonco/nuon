import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  CodeViewer,
  DashboardContent,
  InstallCloudPlatform,
  InstallHistory,
  InstallInputsSection,
  InstallPageSubNav,
  InstallStatuses,
  StatusBadge,
  Section,
  Text,
} from '@/components'
import { getInstall, getInstallRunnerGroup, getOrg } from '@/lib'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, runnerGroup, org] = await Promise.all([
    getInstall({ installId, orgId }),
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
      <Section heading="Runner group">
        {runnerGroup ? (
          <CodeViewer
            initCodeSource={JSON.stringify(runnerGroup, null, 2)}
            language="json"
          />
        ) : (
          <Text>Install runner info here</Text>
        )}
      </Section>
    </DashboardContent>
  )
})
