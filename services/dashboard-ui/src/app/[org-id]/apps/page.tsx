import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  NoApps,
  Loading,
  Notice,
  OrgAppsTable,
  Pagination,
  Section,
} from '@/components'
import { getOrgById, getApps } from '@/lib'
// TODO(nnnat): move segment init script to org dashboard
import type { TApp } from '@/types'
import { SegmentAnalyticsSetOrg, nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Apps | ${org.name} | Nuon`,
  }
}

export default async function Apps({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const { data: org } = await getOrgById({ orgId })

  // Get both apps and account data at page level
  const params_ = new URLSearchParams({
    offset: sp['offset'] || '0',
    limit: '10',
    q: sp['q'] || '',
  }).toString()
  const { data: apps } = await nueQueryData<TApp[]>({
    orgId,
    path: `apps${params_ ? '?' + params_ : params_}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const hasNoApps = !apps || apps.length === 0

  return (
    <>
      {process.env.SEGMENT_WRITE_KEY && <SegmentAnalyticsSetOrg org={org} />}
      <DashboardContent
        breadcrumb={[{ href: `/${orgId}/apps`, text: 'Apps' }]}
      >
        <Section>
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Loading variant="page" loadingText="Loading apps..." />
              }
            >
              <LoadApps
                orgId={orgId}
                offset={sp['offset'] || '0'}
                q={sp['q'] || ''}
              />
            </Suspense>
          </ErrorBoundary>
        </Section>
      </DashboardContent>
    </>
  )
}

const LoadApps: FC<{
  orgId: string
  limit?: number
  offset?: string
  q?: string
}> = async ({ orgId, limit = 10, offset, q }) => {
  const {
    data: apps,
    error,
    headers,
  } = await getApps({
    orgId,
    offset,
    limit,
    q,
  })

  const pageData = {
    hasNext: headers?.['x-nuon-page-next'] || 'false',
    offset: headers?.['x-nuon-page-offset'] || '0',
  }

  return error ? (
    <Notice>Can&apos;t load apps: {error?.error}</Notice>
  ) : apps ? (
    <div className="flex flex-col gap-4 w-full">
      <OrgAppsTable apps={apps} orgId={orgId} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={limit}
      />
    </div>
  ) : (
    <NoApps />
  )
}
