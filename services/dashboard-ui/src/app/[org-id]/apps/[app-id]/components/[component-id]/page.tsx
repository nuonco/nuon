import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  BuildComponentButton,
  ComponentBuildHistory,
  ComponentConfiguration,
  DashboardContent,
  DependentComponents,
  Section,
} from '@/components'
import {
  getApp,
  getAppComponents,
  getComponent,
  getComponentBuilds,
  getComponentConfig,
} from '@/lib'

export default withPageAuthRequired(async function AppComponent({ params }) {
  const appId = params?.['app-id'] as string
  const componentId = params?.['component-id'] as string
  const orgId = params?.['org-id'] as string

  const [app, appComponents, builds, component, componentConfig] =
    await Promise.all([
      getApp({ appId, orgId }),
      getAppComponents({ appId, orgId }),
      getComponentBuilds({ componentId, orgId }),
      getComponent({ componentId, orgId }),
      getComponentConfig({ componentId, orgId }),
    ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}/components`, text: app.name },
        {
          href: `/${orgId}/apps/${app.id}/components/${component.id}`,
          text: component.name,
        },
      ]}
      heading={component.name}
      headingUnderline={component.id}
      statues={
        <BuildComponentButton
          appId={appId}
          componentId={componentId}
          componentName={component?.name}
          orgId={orgId}
        />
      }
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <div className="divide-y flex flex-col flex-auto">
          {component.dependencies && (
            <Section className="flex-initial" heading="Dependencies">
              <DependentComponents
                appId={appId}
                appComponents={appComponents}
                dependentIds={component.dependencies}
                orgId={orgId}
              />
            </Section>
          )}

          <Section heading="Latest config">
            <ComponentConfiguration config={componentConfig} />
          </Section>
        </div>
        <div
          className="border-l overflow-auto lg:min-w-[450px]
lg:max-w-[450px]"
        >
          <Section heading="Build history">
            <ComponentBuildHistory
              appId={appId}
              componentId={componentId}
              initBuilds={builds}
              orgId={orgId}
              shouldPoll
            />
          </Section>
        </div>
      </div>
    </DashboardContent>
  )
})
