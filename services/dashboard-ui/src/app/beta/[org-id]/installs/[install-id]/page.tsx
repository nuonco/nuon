import {
  Card,
  Code,
  Heading,
  InstallEvents,
  InstallSandboxDetails,
  InstallStatus,
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
    <div className="flex flex-col md:flex-row gap-6">
      <div className="flex-auto flex flex-col gap-6">
        <Card>
          <Heading>Most recent event</Heading>
          <div>{events.at(0).operation_name}</div>
        </Card>

        <Card className="max-h-[40rem]">
          <Heading>History</Heading>
          <InstallProvider initInstall={install}>
            <InstallEvents initEvents={events} shouldPoll />
          </InstallProvider>
        </Card>
      </div>

      <div className="flex flex-col gap-6">
        <Card>
          <Heading>Statues</Heading>
          <InstallProvider initInstall={install}>
            <InstallStatus />
          </InstallProvider>
        </Card>

        <Card>
          <Heading>Current inputs</Heading>
          <Code variant="preformated">
            {install.install_inputs.map((input) =>
              JSON.stringify(input.values, null, 2)
            )}
          </Code>
        </Card>

        <Card>
          <Heading>Sandbox config</Heading>
          <InstallProvider initInstall={install}>
            <InstallSandboxDetails />
          </InstallProvider>
        </Card>
      </div>
    </div>
  )
}
