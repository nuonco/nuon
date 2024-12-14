// TODO(nnnat): remove once we have this API change on prod
// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppSandboxConfig,
  AppSandboxVariables,
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
import {
  getInstall,
  getInstallEvents,
  getInstallRunnerGroup,
  getOrg,
} from '@/lib'
import { RUNNERS } from '@/utils'

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
        <Section heading="History" className="overflow-auto history">
          <InstallHistory
            initEvents={events}
            installId={installId}
            orgId={orgId}
            shouldPoll
          />
        </Section>

        <div className="divide-y flex flex-col lg:w-[500px] border-l">
          <Section className="flex-initial" heading="Active sandbox">
            <div className="flex flex-col gap-8">
              <AppSandboxConfig sandboxConfig={install?.app_sandbox_config} />
              <AppSandboxVariables
                variables={install?.app_sandbox_config?.variables}
              />
            </div>
          </Section>

          {install?.install_inputs?.length &&
          install?.install_inputs.some(
            (input) => input.values || input?.redacted_values
          ) ? (
            <InstallInputsSection inputs={install.install_inputs} />
          ) : null}

          {RUNNERS ? (
            <Section className="flex-initial" heading="Runner group">
              <div className="flex flex-col gap-8">
                <Text>{runnerGroup.runners?.length} runners in this group</Text>
                <div className="divide-y">
                  {runnerGroup.runners?.map((runner) => (
                    <div key={runner?.id} className="flex flex-col gap-2">
                      <StatusBadge status={runner?.status} />
                      <Text variant="med-14">{runner?.display_name}</Text>
                    </div>
                  ))}
                </div>
              </div>
            </Section>
          ) : null}

          <Section heading="Cloud platform">
            <InstallCloudPlatform install={install} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
