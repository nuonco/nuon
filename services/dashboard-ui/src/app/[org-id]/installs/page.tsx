import type { Metadata } from 'next'
import { Suspense } from 'react'

import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageLayout } from '@/components/layout/PageLayout'
import { PageContent } from '@/components/layout/PageContent'
import { PageHeader } from '@/components/layout/PageHeader'
import { PageSection } from '@/components/layout/PageSection'
import { getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { Installs } from './installs'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import { DashboardContent, ErrorFallback, Loading, Section } from '@/components'

type TInstallsPageProps = TPageProps<'org-id'>

export async function generateMetadata({
  params,
}: TInstallsPageProps): Promise<Metadata> {
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return {
    title: `Installs | ${org.name} | Nuon`,
  }
}

export default async function InstallsPage({
  params,
  searchParams,
}: TInstallsPageProps) {
  const sp = await searchParams
  const { ['org-id']: orgId } = await params
  const { data: org } = await getOrgById({ orgId })

  return org?.features?.['stratus-layout'] ? (
    <PageLayout
      breadcrumb={{
        baseCrumbs: [
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/installs`,
            text: 'Installs',
          },
        ],
      }}
      isScrollable
    >
      <PageHeader>
        <HeadingGroup>
          <Text variant="h3" weight="stronger" level={1}>
            Installs
          </Text>
          <Text theme="neutral">
            View and manage all deployed installs here.
          </Text>
        </HeadingGroup>
      </PageHeader>
      <PageContent>
        <PageSection>
          <ErrorBoundary
            fallback={
              <Text>
                An error loading your installs, please refresh the page and try
                again.
              </Text>
            }
          >
            <Suspense
              fallback={
                <Loading variant="page" loadingText="Loading installs..." />
              }
            >
              <Installs
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
      breadcrumb={[{ href: `/${orgId}/installs`, text: 'Installs' }]}
    >
      <Section>
        <OldErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading installs..." />
            }
          >
            <Installs
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
            />
          </Suspense>
        </OldErrorBoundary>
      </Section>
    </DashboardContent>
  )
}
