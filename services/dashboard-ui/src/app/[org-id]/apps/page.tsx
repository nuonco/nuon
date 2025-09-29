import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from 'react-error-boundary'
import { DashboardContent, ErrorFallback, Loading, Section } from '@/components'
import { getOrgById } from '@/lib'
// TODO(nnnat): move segment init script to org dashboard
import { SegmentAnalyticsSetOrg } from '@/utils'
import { Apps } from './apps'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Apps | ${org.name} | Nuon`,
  }
}

export default async function AppsPage({ params, searchParams }) {
  const { ['org-id']: orgId } = await params
  const sp = await searchParams
  const { data: org } = await getOrgById({ orgId })

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
              <Apps
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
