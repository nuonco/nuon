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
  InstallManagementDropdown,
  Loading,
  NoComponents,
  Time,
  type TDataInstallComponent,
} from '@/components'
import {
  getComponentBuild,
  getComponent,
  getComponentConfig,
  getInstall,
  getInstallComponents,
} from '@/lib'
import type { TBuild, TComponent } from '@/types'
import { nueQueryData } from "@/utils"

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
      headingMeta={
        <>
          Last updated <Time time={install?.updated_at} format="relative" />
        </>
      }
      statues={
        <div className="flex items-start gap-8">
          <InstallStatuses initInstall={install} shouldPoll />

          <InstallManagementDropdown
            orgId={orgId}
            hasInstallComponents={Boolean(install?.install_components?.length)}
            install={install}
          />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <section className="px-6 py-8">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading loadingText="Loading components..." variant="page" />
            }
          >
            <LoadInstallComponents installId={install?.id} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
})

const LoadInstallComponents: FC<{ installId: string; orgId: string }> = async ({
  installId,
  orgId,
}) => {
  const installComponents = await getInstallComponents({
    installId,
    orgId,
  }).catch(console.error)
  const hydratedInstallComponents =
    installComponents && installComponents?.length
      ? await Promise.all(
          installComponents.map(async (comp, _) => {
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
            const { data } = await nueQueryData<Array<TComponent>>({
              orgId,
              path: `components/${comp.component_id}/dependencies`,
            })

            return {
              ...comp,
              build,
              config,
              deps: data,
            }
          })
        )
      : []

  return hydratedInstallComponents?.length ? (
    <InstallComponentsTable
      installComponents={
        hydratedInstallComponents as Array<TDataInstallComponent>
      }
      installId={installId}
      orgId={orgId}
    />
  ) : (
    <NoComponents />
  )
}
