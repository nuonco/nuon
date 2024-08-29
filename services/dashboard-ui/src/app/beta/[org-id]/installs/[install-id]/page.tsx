import {
  AppSandboxConfig,
  AppSandboxVariables,
  Code,
  Heading,
  InstallHistory,
  InstallInputs,
} from '@/components'
import { InstallProvider } from '@/context'
import { getInstall, getInstallEvents } from '@/lib'

export default async function Install({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, events] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallEvents({ installId, orgId }),
  ])

  return (
    <InstallProvider initInstall={install}>
      <div className="flex flex-col lg:flex-row flex-auto">
        <section className="flex-auto flex flex-col gap-4 px-6 py-8 border-r overflow-auto">
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

          <section className="flex flex-col gap-6 px-6 py-8">
            <Heading>Current inputs</Heading>

            <InstallInputs inputs={install.install_inputs} />
          </section>
        </div>
      </div>
    </InstallProvider>
  )
}
