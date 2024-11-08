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
  InstallStatuses,
  Section,
  SubNav,
  type TLink,
} from '@/components'
import { getInstall, getInstallEvents, getOrg } from '@/lib'

export default withPageAuthRequired(async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/${orgId}/installs/${installId}`, text: 'Status' },
    {
      href: `/${orgId}/installs/${installId}/components`,
      text: 'Components',
    },
  ]

  const [install, events, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallEvents({ installId, orgId }),
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
      meta={<SubNav links={subNavLinks} />}
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

          <Section heading="Cloud platform">
            <InstallCloudPlatform install={install} />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
