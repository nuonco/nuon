import { DashboardContent, DataTable, SubNav, type TLink } from '@/components'
import { getApp, getAppComponents, getOrg } from '@/lib'

export default async function AppComponents({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
    { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
    { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
  ]

  const app = await getApp({ appId, orgId })
  const components = await getAppComponents({ appId, orgId })
  const org = await getOrg({ orgId })

  const tableData = components.reduce((acc, component) => {
    acc.push([
      component.id,
      component.name,
      component.dependencies?.length || 0,
      component?.status,
      component.config_versions,
      `/beta/${orgId}/apps/${appId}/components/${component.id}`,
    ])

    return acc
  }, [])

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
        <DataTable
          headers={['ID', 'Name', 'Dependencies', 'Build', 'Config']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
