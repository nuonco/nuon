import type { Metadata } from 'next'
import { type FC, Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
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
  Pagination,
  Text,
  Time,
} from '@/components'
import { getInstall } from '@/lib'
import type { TInstall, TInstallComponentSummary } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const install = await getInstall({ installId, orgId })

  return {
    title: `${install.name} | Components`,
  }
}

export default async function InstallComponents({ params, searchParams }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
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
          {install?.metadata?.managed_by &&
          install?.metadata?.managed_by === 'nuon/cli/install-config' ? (
            <span className="flex flex-col gap-2">
              <Text isMuted>Managed By</Text>
              <Text>
                <FileCodeIcon />
                Config File
              </Text>
            </span>
          ) : null}
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
              install={install}
              installId={install?.id}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
              types={sp['types'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </section>
    </DashboardContent>
  )
}

const LoadInstallComponents: FC<{
  install: TInstall
  installId: string
  orgId: string
  limit?: string
  offset?: string
  q?: string
  types?: string
}> = async ({ install, installId, orgId, limit = '10', offset, q, types }) => {
  const params = new URLSearchParams({ offset, limit, q, types }).toString()
  const { data, error, headers } = await nueQueryData<
    Array<TInstallComponentSummary>
  >({
    orgId,
    path: `installs/${installId}/components/summary${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return error ? (
    <Notice>Can&apos;t load install components: {error?.error}</Notice>
  ) : data ? (
    <div className="flex flex-col gap-4">
      <InstallComponentsTable
        install={install}
        installComponents={data.sort((a, b) =>
          a?.component_id.localeCompare(b.component_id)
        )}
        installId={installId}
        orgId={orgId}
      />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <NoComponents />
  )
}
