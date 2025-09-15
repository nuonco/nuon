import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { FileCodeIcon } from '@phosphor-icons/react/dist/ssr'
import {
  Link,
  InstallManagementDropdown,
  InstallPageSubNav,
  InstallStatuses,
  InstallActionWorkflowsTable,
  DashboardContent,
  ErrorFallback,
  Loading,
  NoActions,
  Notice,
  Pagination,
  Section,
  Text,
  Time,
} from '@/components'
import { getInstallById } from '@/lib'
import type { TInstallActionWorkflow } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const { data: install } = await getInstallById({ installId, orgId })

  return {
    title: `Actions | ${install.name} | Nuon`,
  }
}

export default async function InstallWorkflowRuns({ params, searchParams }) {
  const { ['org-id']: orgId, ['install-id']: installId } = await params
  const sp = await searchParams
  const { data: install } = await getInstallById({ installId, orgId })

  return (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/installs`, text: 'Installs' },
        { href: `/${orgId}/installs/${install.id}`, text: install.name },
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
          <InstallStatuses />

          <InstallManagementDropdown
            orgId={orgId}
            hasInstallComponents={Boolean(install?.install_components?.length)}
            install={install}
          />
        </div>
      }
      meta={<InstallPageSubNav installId={installId} orgId={orgId} />}
    >
      <Section childrenClassName="flex flex-auto">
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading actions..." />
            }
          >
            <LoadInstallActions
              installId={installId}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
              trigger_types={sp['trigger_types'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

const LoadInstallActions: FC<{
  installId: string
  orgId: string
  limit?: string
  offset?: string
  q?: string
  trigger_types?: string
}> = async ({ installId, orgId, limit = '10', offset, q, trigger_types }) => {
  const params = new URLSearchParams({
    offset,
    limit,
    q,
    trigger_types,
  }).toString()
  const {
    data: actionsWithLatestRun,
    error,
    headers,
  } = await nueQueryData<TInstallActionWorkflow[]>({
    orgId,
    path: `installs/${installId}/action-workflows/latest-runs${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return error ? (
    <Notice className="grow-0 h-max">Can&apos;t load install actions</Notice>
  ) : actionsWithLatestRun ? (
    <div className="flex flex-col gap-4 w-full">
      <InstallActionWorkflowsTable
        actions={actionsWithLatestRun}
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
    <NoActions />
  )
}
