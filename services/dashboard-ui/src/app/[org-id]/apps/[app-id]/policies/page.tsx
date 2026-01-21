import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { PoliciesTable, PoliciesTableSkeleton } from './policies-table'

import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Policies | ${app?.name} | Nuon`,
  }
}

export default async function AppPoliciesPage({ params }: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const [{ data: app }, { data: org }] = await Promise.all([
    getApp({ appId, orgId }),
    getOrg({ orgId }),
  ])

  return org?.features?.['stratus-layout'] ? (
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
            path: `/${orgId}/apps/${appId}/policies`,
            text: 'Policies',
          },
        ]}
      />
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App policies
        </Text>
      </HeadingGroup>

      <div className="flex flex-auto">
        <ErrorBoundary fallback={<>Error loading policies</>}>
          <Suspense fallback={<PoliciesTableSkeleton />}>
            <PoliciesTable appId={appId} orgId={orgId} />
          </Suspense>
        </ErrorBoundary>
      </div>
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app?.id}`, text: app?.name || '' },
        { href: `/${orgId}/apps/${app?.id}/policies`, text: 'Policies' },
      ]}
      heading={app?.name || ''}
      headingUnderline={app?.id}
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <Section childrenClassName="flex flex-auto">
        <OldErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading policies..." />
            }
          >
            <PoliciesTable appId={appId} orgId={orgId} />
          </Suspense>
        </OldErrorBoundary>
      </Section>
    </DashboardContent>
  )
}
