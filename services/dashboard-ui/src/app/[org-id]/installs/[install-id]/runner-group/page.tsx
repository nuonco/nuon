import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  CodeViewer,
  DashboardContent,
  InstallCloudPlatform,
  InstallPageSubNav,
  InstallStatuses,
  StatusBadge,
  Section,
  Text,
} from '@/components'
import {
  getInstall,
  getInstallRunnerGroup,
  getOrg,
  getRunner,
  getRunnerJobs,
} from '@/lib'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, runnerGroup, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallRunnerGroup({ installId, orgId }),
    getOrg({ orgId }),
  ])

  /* const runner = await getRunner({
   *   orgId,
   *   runnerId: runnerGroup?.runners?.at(0)?.id,
   * })
   * const runnerJobs = await getRunnerJobs({
   *   orgId,
   *   runnerId: runnerGroup?.runners?.at(0)?.id,
   * }) */

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
