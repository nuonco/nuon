import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'

import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getOrgById } from '@/lib'
// TODO(nnnat): move segment init script to org dashboard
import { SegmentAnalyticsSetOrg } from '@/lib/segment-analytics'
import { Apps } from './apps'

// old layout components
import { DashboardContent, ErrorFallback, Loading, Section } from '@/components'

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
      {org?.features?.['stratus-layout'] ? (
        <PageLayout
          breadcrumb={{
            baseCrumbs: [
              {
                path: `/${orgId}`,
                text: org?.name,
              },
              {
                path: `/${orgId}/apps`,
                text: 'Apps',
              },
            ],
          }}
          isScrollable
        >
          <PageHeader>
            <HeadingGroup>
              <Text variant="h3" weight="stronger" level={1}>
                Apps
              </Text>
              <Text theme="neutral">Manage your applications here.</Text>
            </HeadingGroup>
          </PageHeader>
          <PageContent>
            <PageSection>
              <ErrorBoundary
                fallback={
                  <Text>
                    An error loading your apps, please refresh the page and try
                    again.
                  </Text>
                }
              >
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
            </PageSection>
          </PageContent>
        </PageLayout>
      ) : (
        <DashboardContent
          breadcrumb={[{ href: `/${orgId}/apps`, text: 'Apps' }]}
        >
          <Section>
            <OldErrorBoundary fallbackRender={ErrorFallback}>
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
            </OldErrorBoundary>
          </Section>
        </DashboardContent>
      )}
    </>
  )
}
