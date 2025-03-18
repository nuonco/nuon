import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  InstallStatuses,
  InstallComponentsTable,
  InstallPageSubNav,
  Loading,
  NoComponents,
  Text,
  Time,
  type TDataInstallComponent,
} from '@/components'
import { InstallManagementDropdown } from '@/components/Installs'
import {
  getComponentBuild,
  getComponent,
  getComponentConfig,
  getInstall,
} from '@/lib'
import type { TBuild, TInstall } from '@/types'
import { USER_REPROVISION } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const installId = params?.['install-id'] as string
  const orgId = params?.['org-id'] as string
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Components`,
  }
}

export default withPageAuthRequired(async function InstallComponents({
  params,
}) {
  const orgId = params?.['org-id'] as string
  const installId = params?.['install-id'] as string
  const install = await getInstall({ orgId, installId })

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
              orgId={orgId}
              hasInstallComponents={Boolean(
                install?.install_components?.length
              )}
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
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading loadingText="Loading components..." variant="page" />
            }
          >
            <LoadInstallComponents install={install} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
})

const LoadInstallComponents: FC<{ install: TInstall; orgId: string }> = async ({
  install,
  orgId,
}) => {
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

  return hydratedInstallComponents?.length ? (
    <InstallComponentsTable
      installComponents={
        hydratedInstallComponents as Array<TDataInstallComponent>
      }
      installId={install.id}
      orgId={orgId}
    />
  ) : (
    <NoComponents />
  )
}
