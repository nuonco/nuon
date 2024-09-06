import {
  ComponentBuildHistory,
  ComponentConfiguration,
  DashboardContent,
  DependentComponents,
  Heading,
  Text,
} from '@/components'
import {
  getApp,
  getAppComponents,
  getComponent,
  getComponentBuilds,
  getComponentConfig,
  getOrg,
} from '@/lib'

export default async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string

  const [app, appComponents, builds, component, componentConfig, org] =
    await Promise.all([
      getApp({ appId, orgId }),
      getAppComponents({ appId, orgId }),
      getComponentBuilds({ componentId, orgId }),
      getComponent({ componentId, orgId }),
      getComponentConfig({ componentId, orgId }),
      getOrg({ orgId }),
    ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/apps`, text: 'Apps' },
        { href: `/beta/${org.id}/apps/${app.id}`, text: app.name },
        {
          href: `/beta/${org.id}/apps/${app.id}/components/${component.id}`,
          text: component.name,
        },
      ]}
      heading={component.name}
      headingUnderline={component.id}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <div className="divide-y flex flex-col flex-auto">
          {component.dependencies && (
            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Dependencies</Heading>

              <DependentComponents
                appId={appId}
                appComponents={appComponents}
                dependentIds={component.dependencies}
                orgId={orgId}
              />
            </section>
          )}

          <section className="flex flex-col gap-6 px-6 py-8">
            <Heading>Latest config</Heading>

            <ComponentConfiguration config={componentConfig} />
          </section>
        </div>
        <section
          className="flex flex-col gap-4 px-6 py-8 border-l overflow-auto lg:min-w-[500px]
lg:max-w-[500px]"
        >
          <Heading>Build history</Heading>

          <ComponentBuildHistory
            appId={appId}
            componentId={componentId}
            initBuilds={builds}
            orgId={orgId}
            shouldPoll
          />
        </section>
      </div>
    </DashboardContent>
  )
}
