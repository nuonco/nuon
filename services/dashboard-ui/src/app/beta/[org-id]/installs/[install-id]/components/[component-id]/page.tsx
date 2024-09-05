import {
  ComponentConfiguration,
  DashboardContent,
  DependentComponents,
  Heading,
  InstallComponentDeploys,
  Link,
} from '@/components'
import {
  getAppComponents,
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponent,
  getOrg,
} from '@/lib'

export default async function InstallComponent({ params }) {
  const installComponentId = params?.['component-id'] as string
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string

  const [install, installComponent, org] = await Promise.all([
    getInstall({ installId, orgId }),
    getInstallComponent({
      installComponentId,
      installId,
      orgId,
    }),
    getOrg({ orgId }),
  ])

  const [appComponents, component, componentConfig] = await Promise.all([
    getAppComponents({ appId: installComponent?.component?.app_id, orgId }),
    getComponent({ componentId: installComponent.component_id, orgId }),
    getComponentConfig({ componentId: installComponent?.component_id, orgId }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/installs`, text: 'Installs' },
        { href: `/beta/${org.id}/installs/${install.id}`, text: install.name },
        {
          href: `/beta/${org.id}/installs/${install.id}/components/${installComponent.id}`,
          text: component.name,
        },
      ]}
      heading={component.name}
      headingUnderline={installComponent.id}
    >
      <div className="flex flex-col lg:flex-row flex-auto">
        <div className="divide-y flex-auto  flex flex-col overlfow-auto">
          {component.dependencies && (
            <section className="flex flex-col gap-6 px-6 py-8">
              <Heading>Dependencies</Heading>

              <DependentComponents
                appComponents={appComponents}
                dependentIds={component.dependencies}
              />
            </section>
          )}

          <section className="flex flex-col gap-6 px-6 py-8">
            <Heading>Latest config</Heading>

            <ComponentConfiguration config={componentConfig} />
          </section>
        </div>
        <section className="flex flex-col gap-4 px-6 py-8 border-l overflow-auto lg:min-w-[500px] lg:max-w-[500px]">
          <Heading>Deploy history</Heading>
          <InstallComponentDeploys
            initDeploys={installComponent.install_deploys}
            installId={installId}
            installComponentId={installComponentId}
            orgId={orgId}
          />
        </section>
      </div>
    </DashboardContent>
  )
}
