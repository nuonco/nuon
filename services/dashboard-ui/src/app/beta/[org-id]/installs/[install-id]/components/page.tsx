import {
  DashboardContent,
  InstallStatus,
  InstallComponentsTable,
  SubNav,
  type TDataInstallComponent,
  type TLink,
} from '@/components'
import { InstallProvider } from '@/context'
import {
  getInstall,
  getBuild,
  getComponent,
  getComponentConfig,
  getOrg,
} from '@/lib'

export default async function InstallComponents({ params }) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const subNavLinks: Array<TLink> = [
    { href: `/beta/${orgId}/installs/${installId}`, text: 'Status' },
    {
      href: `/beta/${orgId}/installs/${installId}/components`,
      text: 'Components',
    },
  ]

  const [install, org] = await Promise.all([
    getInstall({ orgId, installId }),
    getOrg({ orgId }),
  ])

  const hydratedInstallComponents = await Promise.all(
    install.install_components.map(async (comp, _, arr) => {
      const build = await getBuild({
        buildId: comp.install_deploys?.[0]?.build_id,
        orgId,
      })
      const config = await getComponentConfig({
        componentId: comp.component.id,
        componentConfigId: build.component_config_connection_id,
        orgId,
      })
      const appComponent = await getComponent({
        componentId: comp.component_id,
        orgId,
      })
      const deps = arr.filter((c) =>
        appComponent.dependencies?.some((d) => d === c.component_id)
      )

      return {
        ...comp,
        build,
        config,
        deps,
      }
    })
  )

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/beta/${org.id}`, text: org.name },
        { href: `/beta/${org.id}/installs`, text: 'Installs' },
        { href: `/beta/${org.id}/installs/${install.id}`, text: install.name },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={
        <div>
          <InstallProvider initInstall={install}>
            <InstallStatus />
          </InstallProvider>
        </div>
      }
      meta={<SubNav links={subNavLinks} />}
    >
      <section className="px-6 py-8">
        <InstallComponentsTable
          installComponents={
            hydratedInstallComponents as Array<TDataInstallComponent>
          }
          installId={installId}
          orgId={orgId}
        />
      </section>
    </DashboardContent>
  )
}
