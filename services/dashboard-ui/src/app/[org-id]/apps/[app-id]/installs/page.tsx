import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getOrg } from '@/lib'
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
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Installs | ${app.name} | Nuon`,
  }
}

export default async function AppInstallsPage({ params, searchParams }) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [{ data: app, error }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getOrg({ orgId }),
  ])

  if (error) {
    notFound()
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name,
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name,
          },
          {
            path: `/${orgId}/apps/${appId}/installs`,
            text: 'Installs',
          },
        ]}
      />
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
  )
}
