import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  InstallStatuses,
  InstallComponentsTable,
  InstallManagementDropdown,
  InstallPageSubNav,
  NoComponents,
  Text,
  Time,
  type TDataInstallComponent,
} from '@/components'
import {
  getAppLatestInputConfig,
  getComponentBuild,
  getComponent,
  getComponentConfig,
  getInstall,
} from '@/lib'
import type { TBuild } from '@/types'
import { USER_REPROVISION, INSTALL_UPDATE } from '@/utils'

export default withPageAuthRequired(async function InstallComponents({
  params,
}) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const [install] = await Promise.all([getInstall({ orgId, installId })])

  const appInputConfigs =
    (await getAppLatestInputConfig({
      appId: install?.app_id,
      orgId,
    }).catch(console.error)) || undefined

  const hydratedInstallComponents =
    install.install_components && install.install_components?.length
      ? await Promise.all(
          install.install_components.map(async (comp, _, arr) => {
            const build = await getComponentBuild({
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
        { href: `/${orgId}/installs`, text: 'Installs' },
        {
          href: `/${orgId}/installs/${install.id}`,
          text: install.name,
        },
      ]}
      heading={install.name}
      headingUnderline={install.id}
      statues={
        <div className="flex items-start gap-8">
          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Created
            </Text>
            <Time variant="reg-12" time={install?.created_at} />
          </span>

          <span className="flex flex-col gap-2">
            <Text className="text-cool-grey-600 dark:text-cool-grey-500">
              Updated
            </Text>
            <Time variant="reg-12" time={install?.updated_at} />
          </span>
          <InstallStatuses initInstall={install} shouldPoll />
          {USER_REPROVISION ? (
            <InstallManagementDropdown
              installId={installId}
              orgId={orgId}
              hasInstallComponents={Boolean(
                install?.install_components?.length
              )}
              hasUpdateInstall={INSTALL_UPDATE}
              inputConfig={appInputConfigs}
              install={install}
            />
          ) : null}
        </div>
      }
      meta={
        <InstallPageSubNav
          installId={installId}
          orgId={orgId}
          runnerId={install?.runner_id}
        />
      }
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
