// TODO(nnnat): remove once we have this API changes on prod
// @ts-nocheck
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppSandboxConfig,
  AppSandboxVariables,
  DashboardContent,
  Heading,
  InstallCloudPlatform,
  InstallCloudPlatformDetails,
  InstallHistory,
  InstallInputs,
  InstallStatus,
  InstallStatuesV2,
  SubNav,
  type TLink,
} from '@/components'
import { InstallProvider } from '@/context'
import { getInstall, getInstallEvents, getOrg } from '@/lib'

export default withPageAuthRequired(
  async function Install({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string
    const subNavLinks: Array<TLink> = [
      { href: `/beta/${orgId}/installs/${installId}`, text: 'Status' },
      {
        href: `/beta/${orgId}/installs/${installId}/components`,
        text: 'Components',
      },
    ]

    const [install, events, org] = await Promise.all([
      getInstall({ installId, orgId }),
      getInstallEvents({ installId, orgId }),
      getOrg({ orgId }),
    ])

    console.log('install', install)

    return (
      <DashboardContent
        breadcrumb={[
          { href: `/beta/${org.id}/apps`, text: org.name },
          { href: `/beta/${org.id}/installs`, text: 'Installs' },
          {
            href: `/beta/${org.id}/installs/${install.id}`,
            text: install.name,
          },
        ]}
        heading={install.name}
        headingUnderline={install.id}
        statues={<InstallStatuesV2 install={install} />}
        meta={<SubNav links={subNavLinks} />}
      >
        <div className="flex flex-col lg:flex-row flex-auto">
          <section className="flex-auto flex flex-col gap-4 px-6 py-8 border-r overflow-auto history">
            <Heading>History</Heading>

            <InstallHistory
              initEvents={events}
              installId={installId}
              orgId={orgId}
              shouldPoll
            />
          </section>

          <div className="divide-y flex flex-col lg:w-[550px]">
            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Active sandbox</Heading>

              <AppSandboxConfig sandboxConfig={install?.app_sandbox_config} />
              <AppSandboxVariables
                variables={install?.app_sandbox_config?.variables}
              />
            </section>

            {install?.install_inputs?.length &&
              install?.install_inputs.some(
                (input) => input.values || input?.redacted_values
              ) && (
                <section className="flex flex-col gap-6 px-6 py-8">
                  <Heading>Current inputs</Heading>

                  <InstallInputs inputs={install.install_inputs} />
                </section>
              )}

            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Cloud platform</Heading>

              <InstallCloudPlatform install={install} />
            </section>
          </div>
        </div>
      </DashboardContent>
    )
  },
  { returnTo: '/dashboard' }
)
