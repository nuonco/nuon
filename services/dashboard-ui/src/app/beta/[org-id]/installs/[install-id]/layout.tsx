import {
  DashboardContent,
  Heading,
  InstallStatus,
  Text,
  SubNav,
  type TLink,
} from '@/components'
import { InstallProvider } from '@/context'
import { getInstall, getOrg } from '@/lib'

export default async function InstallLayout({ children, params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/installs/${installId}`, text: 'Status' },
    {
      href: `/beta/${orgId}/installs/${installId}/components`,
      text: 'Components',
    },
  ]
  const [install, org] = await Promise.all([
    getInstall({ orgId, installId }),
    getOrg({ orgId }),
  ])

  return (
    <DashboardContent breadcrumb={[org.name, 'Installs', install.name]}>
      <>
        <header className="px-6 pt-8 flex flex-col pt-6 gap-6 border-b">
          <div className="flex items-center justify-between">
            <hgroup className="flex flex-col gap-2">
              <Heading>{install.name}</Heading>
              <Text className="font-mono" variant="overline">
                {install.id}
              </Text>
            </hgroup>

            <div>
              <InstallProvider initInstall={install}>
                <InstallStatus />
              </InstallProvider>
            </div>
          </div>

          <SubNav links={subNavLinks} />
        </header>

        {children}
      </>
    </DashboardContent>
  )
}
