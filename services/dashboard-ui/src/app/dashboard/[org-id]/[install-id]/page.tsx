import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  Heading,
  InstallComponentsListCard,
  InstallEventsCard,
  InstallStatus,
  InstallPlatformType,
  InstallCloudPlatformDetailsCard,
  InstallSandboxDetailsCard,
  InstallRegion,
  Page,
  PageHeader,
  PageSummary,
  PageTitle,
  Text,
} from '@/components'
import { InstallProvider } from '@/context'
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
      <InstallProvider initInstall={install} shouldPoll>
        <Page
          header={
            <PageHeader
              info={<InstallStatus />}
              title={<PageTitle overline={installId} title={install.name} />}
              summary={
                <PageSummary>
                  <Text variant="status">{install.app?.name}</Text>
                  <InstallPlatformType isIconOnly />
                  <InstallRegion />
                </PageSummary>
              }
            />
          }
          links={[{ href: orgId }, { href: installId }]}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 w-full h-fit">
            <div className="flex flex-col gap-6">
              <Heading variant="subtitle">History</Heading>
              <InstallEventsCard initEvents={events} shouldPoll />
            </div>

            <div className="flex flex-col gap-6">
              <Heading variant="subtitle">Components</Heading>
              <InstallComponentsListCard className="max-h-[40rem]" />
            </div>

            <div className="flex flex-col gap-6">
              <Heading variant="subtitle">Details</Heading>
              <InstallSandboxDetailsCard />
              <InstallCloudPlatformDetailsCard />
            </div>
          </div>
        </Page>
      </InstallProvider>
    )
  },
  { returnTo: '/dashboard' }
)
