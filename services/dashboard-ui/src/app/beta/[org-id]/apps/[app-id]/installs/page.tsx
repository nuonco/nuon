import {
  AppInstallsTable,
  DashboardContent,
  SubNav,
  type TLink,
} from '@/components'
import { getApp, getAppInstalls, getOrg } from '@/lib'

export default async function AppInstalls({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
    { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
    { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
  ]
  const [app, installs, org] = await Promise.all([
    getApp({ appId, orgId }),
    getAppInstalls({ appId, orgId }),
    getOrg({ orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/apps`, text: 'Apps' },
        { href: `/beta/${org.id}/apps/${app.id}`, text: app.name },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      meta={<SubNav links={subNavLinks} />}
    >
      <section className="px-6 py-8">
        <AppInstallsTable
          installs={installs.map((install) => ({ ...install, app }))}
          orgId={orgId}
        />
      </section>
    </DashboardContent>
  )
}
