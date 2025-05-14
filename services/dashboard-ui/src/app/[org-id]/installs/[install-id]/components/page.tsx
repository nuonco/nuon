import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { withPageAuthRequired } from '@auth0/nextjs-auth0'
import {
  DashboardContent,
  ErrorFallback,
  Link,
  InstallStatuses,
  InstallComponentsTable,
  InstallPageSubNav,
  InstallManagementDropdown,
  Loading,
  NoComponents,
  Text,
  Time,
  type TDataInstallComponent,
} from '@/components'
import {
  getAppComponents,
  getComponentBuild,
  getComponentConfig,
  getInstall,
  getInstallComponents,
} from '@/lib'
import type { TBuild, TComponent, TInstall } from '@/types'
import { nueQueryData } from '@/utils'

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
          href: `/${orgId}/installs/${install?.id}`,
          text: install?.name,
        },
      ]}
      heading={install?.name}
      headingUnderline={install?.id}
      headingMeta={
        <>
          Last updated <Time time={install?.updated_at} format="relative" />
        </>
      }
      statues={
        <div className="flex items-start gap-8">
          <span className="flex flex-col gap-2">
            <Text isMuted>App config</Text>
            <Text>
              <Link href={`/${orgId}/apps/${install.app_id}`}>
                {install?.app?.name}
              </Link>
            </Text>
          </span>
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
            <LoadInstallComponents
              appId={install?.app_id}
              installId={install?.id}
              orgId={orgId}
            />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
})

const LoadInstallComponents: FC<{
  appId: string
  installId: string
  orgId: string
}> = async ({ appId, installId, orgId }) => {
  const installComponents = await getInstallComponents({
    installId,
    orgId,
  }).catch(console.error)
  const appComponents = await getAppComponents({
    appId,
    orgId,
  }).catch(console.error)

  const hydratedInstallComponents =
    installComponents && installComponents?.length && appComponents
      ? await Promise.all(
          installComponents?.map(async (ic) => {
            const config = await getComponentConfig({
              componentId: ic?.component_id,
              orgId,
            })

            return {
              ...ic,
              config,
              deps: appComponents.filter((c) =>
                config?.component_dependency_ids?.some((d) => d === c.id)
              ),
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
