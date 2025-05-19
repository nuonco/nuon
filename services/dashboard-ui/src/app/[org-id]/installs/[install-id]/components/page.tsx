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
  Notice,
  Text,
  Time,
  type TTableInstallComponent,
} from '@/components'
import { getComponentConfig, getInstall } from '@/lib'
import type { TInstallComponentSummary } from '@/types'
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
        {
          href: `/${orgId}/installs/${install?.id}/components`,
          text: 'Components',
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
            <LoadInstallComponents installId={install?.id} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
})

const LoadInstallComponents: FC<{
  installId: string
  orgId: string
}> = async ({ installId, orgId }) => {
  const { data, error } = await nueQueryData<Array<TInstallComponentSummary>>({
    orgId,
    path: `installs/${installId}/components/summary`,
  })

  const hydratedInstallComponents = data
    ? await Promise.all(
        data?.map(async (ic) => {
          const config = await getComponentConfig({
            componentId: ic?.component_id,
            orgId,
          }).catch(console.error)

          return {
            ...ic,
            config,
          }
        })
      )
    : []

  return error ? (
    <Notice>Can&apos;t load install components: {error?.error}</Notice>
  ) : hydratedInstallComponents?.length ? (
    <InstallComponentsTable
      installComponents={
        hydratedInstallComponents as Array<TTableInstallComponent>
      }
      installId={installId}
      orgId={orgId}
    />
  ) : (
    <NoComponents />
  )
}
