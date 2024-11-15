import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  InstallStatuses,
  InstallComponentsTable,
  InstallPageSubNav,
  NoComponents,
  type TDataInstallComponent,
} from '@/components'
import {
  getInstall,
  getBuild,
  getComponent,
  getComponentConfig,
  getOrg,
} from '@/lib'
import type { TBuild } from '@/types'

export default withPageAuthRequired(async function InstallComponents({
  params,
}) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install, org] = await Promise.all([
    getInstall({ orgId, installId }),
    getOrg({ orgId }),
  ])

  const hydratedInstallComponents =
    install.install_components && install.install_components?.length
      ? await Promise.all(
          install.install_components.map(async (comp, _, arr) => {
            const build = await getBuild({
              buildId: comp.install_deploys?.[0]?.build_id,
              orgId,
            }).catch((err) => console.error(err))
            const config = await getComponentConfig({
              componentId: comp.component.id,
              componentConfigId: (build as TBuild)
                ?.component_config_connection_id,
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
      : []

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${org.id}/apps`, text: org.name },
        { href: `/${org.id}/installs`, text: 'Installs' },
        {
          href: `/${org.id}/installs/${install.id}`,
          text: install.name,
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={<InstallStatuses initInstall={install} shouldPoll />}
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <section className="px-6 py-8">
        {hydratedInstallComponents?.length ? (
          <InstallComponentsTable
            installComponents={
              hydratedInstallComponents as Array<TDataInstallComponent>
            }
            installId={installId}
            orgId={orgId}
          />
        ) : (
          <NoComponents />
        )}
      </section>
    </DashboardContent>
  )
})
