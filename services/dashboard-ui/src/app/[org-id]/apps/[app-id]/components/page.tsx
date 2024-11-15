import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  AppComponentsTable,
  AppPageSubNav,
  DashboardContent,
  NoComponents,
} from '@/components'
import {
  getApp,
  getAppComponents,
  getComponentBuilds,
  getComponentConfig,
  getOrg,
} from '@/lib'

export default withPageAuthRequired(async function AppComponents({ params }) {
  const appId = params?.['app-id'] as string
  const orgId = params?.['org-id'] as string
  const [app, components, org] = await Promise.all([
    getApp({ appId, orgId }),
    getAppComponents({ appId, orgId }),
    getOrg({ orgId }),
  ])
  const hydratedComponents = await Promise.all(
    components.map(async (comp, _, arr) => {
      const [config, builds] = await Promise.all([
        getComponentConfig({ componentId: comp.id, orgId }),
        getComponentBuilds({ componentId: comp.id, orgId }),
      ])
      const deps = arr.filter((c) => comp.dependencies?.some((d) => d === c.id))

      return {
        ...comp,
        config,
        deps,
        latestBuild: builds[0],
      }
    })
  )

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/apps`, text: 'Apps' },
        { href: `/${org.id}/apps/${app.id}`, text: app.name },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <section className="px-6 py-8">
        {components.length ? (
          <AppComponentsTable
            components={hydratedComponents}
            appId={appId}
            orgId={orgId}
          />
        ) : (
          <NoComponents />
        )}
      </section>
    </DashboardContent>
  )
})
