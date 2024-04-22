import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Card,
  EventsTimeline,
  Heading,
  InstallComponents,
  InstallPageHeader,
  CloudDetails,
  Page,
  SandboxDetails,
} from '@/components'
import { getInstall, getInstallEvents } from '@/lib'

export default withPageAuthRequired(
  async function InstallDashboard({ params }) {
    const orgId = params?.['org-id'] as string
    const installId = params?.['install-id'] as string

    const [install, events] = await Promise.all([
      getInstall({ installId, orgId }),
      getInstallEvents({ installId, orgId }),
    ])

    return (
      <Page
        header={<InstallPageHeader {...install} />}
        links={[{ href: install?.org_id }, { href: install?.id }]}
      >
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 w-full h-fit">
          <div className="flex flex-col gap-6">
            <Heading variant="subtitle">History</Heading>
            <Card className="max-h-[40rem]">
              <EventsTimeline
                feedId={install?.id}
                orgId={install?.org_id}
                initEvents={events}
              />
            </Card>
          </div>

          <div className="flex flex-col gap-6">
            <Heading variant="subtitle">Components</Heading>
            <Card className="max-h-[40rem]">
              <InstallComponents components={install?.install_components} />
            </Card>
          </div>

          <div className="flex flex-col gap-6">
            <Heading variant="subtitle">Details</Heading>

            <Card>
              <SandboxDetails {...install?.app_sandbox_config} />
            </Card>

            <Card>
              <CloudDetails {...install} />
            </Card>
          </div>
        </div>
      </Page>
    )
  },
  { returnTo: '/dashboard' }
)
