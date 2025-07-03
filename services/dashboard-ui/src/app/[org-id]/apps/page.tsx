import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  NoApps,
  Loading,
  OrgAppsTable,
  Pagination,
  Section,
} from '@/components'
import { getOrg } from '@/lib'
// TODO(nnnat): move segment init script to org dashboard
import type { TApp } from '@/types'
import { SegmentAnalyticsSetOrg, nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const org = await getOrg({ orgId })

  return {
    title: `${org.name} | Apps`,
  }
}

export default async function Apps({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const org = await getOrg({ orgId })

  return (
    <>
      {process.env.SEGMENT_WRITE_KEY && <SegmentAnalyticsSetOrg org={org} />}
      <DashboardContent breadcrumb={[{ href: `/${orgId}/apps`, text: 'Apps' }]}>
        <Section>
          <ErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={
                <Loading variant="page" loadingText="Loading apps..." />
              }
            >
              <LoadApps orgId={orgId} offset={sp['offset'] || '0'} />
            </Suspense>
          </ErrorBoundary>
        </Section>
      </DashboardContent>
    </>
  )
}

const LoadApps: FC<{
  orgId: string
  limit?: string
  offset?: string
}> = async ({ orgId, limit = '10', offset }) => {
  const params = new URLSearchParams({ offset, limit }).toString()
  const {
    data: apps,
    error,
    headers,
  } = await nueQueryData<TApp[]>({
    orgId,
    path: `apps${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return apps?.length && !error ? (
    <div className="flex flex-col gap-4 w-full">
      <OrgAppsTable apps={apps} orgId={orgId} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <NoApps />
  )
}
