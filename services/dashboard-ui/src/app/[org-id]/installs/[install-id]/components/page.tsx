import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
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
} from '@/components'
import { getInstall } from '@/lib'
import type { TInstallComponentSummary } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Components`,
  }
}

export default async function InstallComponents({ params }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
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
}

const LoadInstallComponents: FC<{
  installId: string
  orgId: string
}> = async ({ installId, orgId }) => {
  const { data, error } = await nueQueryData<Array<TInstallComponentSummary>>({
    orgId,
    path: `installs/${installId}/components/summary`,
  })

  return error ? (
    <Notice>Can&apos;t load install components: {error?.error}</Notice>
  ) : data?.length ? (
    <InstallComponentsTable
      installComponents={data.sort((a, b) =>
        a?.component_id.localeCompare(b.component_id)
      )}
      installId={installId}
      orgId={orgId}
    />
  ) : (
    <NoComponents />
  )
}
