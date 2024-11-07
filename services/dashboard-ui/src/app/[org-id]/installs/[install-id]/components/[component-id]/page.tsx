import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  ComponentConfiguration,
  DashboardContent,
  DependentComponents,
  Heading,
  InstallComponentDeploys,
} from '@/components'
import {
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponent,
  getOrg,
} from '@/lib'
import type { TInstallComponent } from '@/types'

export default withPageAuthRequired(async function InstallComponent({
  params,
}) {
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

  const [component, componentConfig] = await Promise.all([
    getComponent({ componentId: installComponent.component_id, orgId }),
    getComponentConfig({
      componentId: installComponent?.component_id,
      orgId,
    }),
  ])

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}/components`,
          text: install.name,
        },
        {
          href: `/${org.id}/installs/${install.id}/components/${installComponent.id}`,
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
                dependentIds={component.dependencies}
                installComponents={
                  install?.install_components as Array<TInstallComponent>
                }
                installId={installId}
                orgId={orgId}
              />
            </section>
          )}

          <section className="flex flex-col gap-6 px-6 py-8">
            <Heading>Component config</Heading>

            <ComponentConfiguration config={componentConfig} />
          </section>
        </div>
        <section className="flex flex-col gap-4 px-6 py-8 border-l overflow-auto lg:min-w-[450px] lg:max-w-[450px]">
          <Heading>Deploy history</Heading>
          <InstallComponentDeploys
            component={component}
            initDeploys={installComponent.install_deploys}
            installId={installId}
            installComponentId={installComponent.id}
            orgId={orgId}
            shouldPoll
          />
        </section>
      </div>
    </DashboardContent>
  )
})
