import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppConfigs, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { ComponentsTable, ComponentsTableSkeleton } from './components-table'

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
import { AppComponents } from './components'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Components | ${app.name} | Nuon`,
  }
}

export default async function AppComponentsPage({
  params,
  searchParams,
}: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [{ data: app }, { data: configs }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getAppConfigs({ appId, orgId }),
    getOrg({ orgId }),
  ])

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
            path: `/${orgId}/apps/${appId}/components`,
            text: 'Components',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App components
        </Text>
      </HeadingGroup>

      {/* old layout stuff */}
      <div className="flex flex-auto">
        <ErrorBoundary fallback={<>Error with components</>}>
          <Suspense fallback={<ComponentsTableSkeleton />}>
            <ComponentsTable
              appId={appId}
              orgId={orgId}
              offset={sp['offset'] || '0'}
              q={sp['q'] || ''}
              types={sp['types'] || ''}
            />
          </Suspense>
        </ErrorBoundary>
      </div>
      {/* old layout stuff */}
    </PageSection>
  )
}
