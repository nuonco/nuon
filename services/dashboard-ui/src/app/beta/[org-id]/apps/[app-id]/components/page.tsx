import {
  AppComponentsTable,
  DashboardContent,
  SubNav,
  type TLink,
} from '@/components'
import { getApp, getAppComponents, getComponentConfig, getOrg } from '@/lib'

export default async function AppComponents({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/apps/${appId}`, text: 'Config' },
    { href: `/beta/${orgId}/apps/${appId}/components`, text: 'Components' },
    { href: `/beta/${orgId}/apps/${appId}/installs`, text: 'Installs' },
  ]

  const [app, components, org] = await Promise.all([
    getApp({ appId, orgId }),
    getAppComponents({ appId, orgId }),
    getOrg({ orgId }),
  ])
  const hydratedComponents = await Promise.all(
    components.map(async (comp, _, arr) => {
      const config = await getComponentConfig({ componentId: comp.id, orgId })
      const deps = arr.filter((c) => comp.dependencies?.some((d) => d === c.id))

      return {
        ...comp,
        config,
        deps,
      }
    })
  )

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
        <AppComponentsTable
          components={hydratedComponents}
          appId={appId}
          orgId={orgId}
        />
      </section>
    </DashboardContent>
  )
}
