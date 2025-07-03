import type { Metadata } from 'next'
import { Suspense, type FC } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import {
  DashboardContent,
  ErrorFallback,
  NoInstalls,
  OrgInstallsTable,
  Loading,
  Pagination,
  Section,
} from '@/components'
import { getOrg } from '@/lib'
import type { TInstall } from '@/types'
import { nueQueryData } from '@/utils'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const org = await getOrg({ orgId })

  return {
    title: `${org.name} | Installs`,
  }
}

export default async function Installs({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  return (
    <DashboardContent
      breadcrumb={[{ href: `/${orgId}/installs`, text: 'Installs' }]}
    >
      <Section>
        <ErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading installs..." />
            }
          >
            <LoadInstalls orgId={orgId} offset={sp['offset'] || '0'} />
          </Suspense>
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}

const LoadInstalls: FC<{
  orgId: string
  limit?: string
  offset?: string
}> = async ({ orgId, limit = '10', offset }) => {
  const params = new URLSearchParams({ offset, limit }).toString()
  const {
    data: installs,
    error,
    headers,
  } = await nueQueryData<TInstall[]>({
    orgId,
    path: `installs${params ? '?' + params : params}`,
    headers: {
      'x-nuon-pagination-enabled': true,
    },
  })

  const pageData = {
    hasNext: headers?.get('x-nuon-page-next') || 'false',
    offset: headers?.get('x-nuon-page-offset') || '0',
  }

  return installs?.length && !error ? (
    <div className="flex flex-col gap-4 w-full">
      <OrgInstallsTable orgId={orgId} installs={installs} />
      <Pagination
        param="offset"
        pageData={pageData}
        position="center"
        limit={parseInt(limit)}
      />
    </div>
  ) : (
    <NoInstalls />
  )
}
