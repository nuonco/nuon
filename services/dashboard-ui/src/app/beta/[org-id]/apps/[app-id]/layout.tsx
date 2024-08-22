import {
  DashboardContent,
  Heading,
  Text,
  SubNav,
  type TLink,
} from '@/components'
import { getApp, getOrg } from '@/lib'

export default async function AppLayout({ children, params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
    { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
    { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
  ]
  const [org, app] = await Promise.all([
    getOrg({ orgId }),
    getApp({ appId, orgId }),
  ])

  return (
    <DashboardContent breadcrumb={[org.name, 'Apps', app.name]}>
      <>
        <header className="px-6 pt-8 flex flex-col pt-6 gap-6 border-b">
          <div className="flex items-center justify-between">
            <hgroup className="flex flex-col gap-2">
              <Heading>{app.name}</Heading>
              <Text className="font-mono" variant="overline">
                {app.id}
              </Text>
            </hgroup>
          </div>

          <SubNav links={subNavLinks} />
        </header>

        {children}
      </>
    </DashboardContent>
  )
}
