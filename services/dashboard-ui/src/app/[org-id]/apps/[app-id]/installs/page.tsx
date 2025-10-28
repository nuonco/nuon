import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getAppById, getOrgById } from '@/lib'
import { InstallsTable, InstallsTableSkeleton } from './installs-table'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'
import { AppInstalls } from './installs'

export async function generateMetadata({ params }): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getAppById({ appId, orgId })

  return {
    title: `Installs | ${app.name} | Nuon`,
  }
}

export default async function AppInstallsPage({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [{ data: app, error }, { data: org }] = await Promise.all([
    getAppById({ appId, orgId }),
    getOrgById({ orgId }),
  ])

  if (error) {
    notFound()
  }

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App installs
        </Text>
      </HeadingGroup>

      {/* old layout stuff */}
      <ErrorBoundary fallback={<>Error loading app installs</>}>
        <Suspense fallback={<InstallsTableSkeleton />}>
          <InstallsTable
            appId={appId}
            orgId={orgId}
            offset={sp['offset'] || '0'}
            q={sp['q'] || ''}
          />
        </Suspense>
      </ErrorBoundary>
      {/* old layout stuff */}
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
        { href: `/${orgId}/apps/${app.id}/installs`, text: 'Installs' },
      ]}
      heading={app.name}
      headingUnderline={app.id}
      statues={
        app?.cloud_platform === 'aws' || app.cloud_platform === 'azure' ? (
          <AppCreateInstallButton platform={app?.cloud_platform} />
        ) : null
      }
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <Section>
        <OldErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading installs..." />
            }
          >
            <AppInstalls
              app={app}
              appId={appId}
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
