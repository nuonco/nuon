import {
  ComponentConfigType,
  DashboardContent,
  DataTable,
  Heading,
  Status,
  SubNav,
  Text,
  type TLink,
} from '@/components'
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
    /* eslint react/jsx-key: 0 */
    acc.push([
      <div className="flex flex-col gap-2">
        <Heading variant="subheading">{component?.name}</Heading>
        <Text variant="caption">{component.id}</Text>
      </div>,
      <Text variant="caption">
        <ComponentConfigType componentId={component.id} orgId={orgId} />
      </Text>,
      <Text variant="caption">{component.dependencies?.length || 0}</Text>,
      <Status status={component?.status} />,
      <Text variant="caption">{component.config_versions}</Text>,
      `/beta/${orgId}/apps/${appId}/components/${component.id}`,
    ])
    /* eslint react/jsx-key: 1 */
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
          headers={['Name', 'Type', 'Dependencies', 'Build', 'Config']}
          initData={tableData}
        />
      </section>
    </DashboardContent>
  )
}
