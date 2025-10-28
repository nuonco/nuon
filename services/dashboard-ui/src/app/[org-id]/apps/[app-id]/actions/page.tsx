import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getAppById, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { ActionsTable, ActionsTableSkeleton } from './actions-table'

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
import { AppActions } from './actions'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getAppById({ appId, orgId })

  return {
    title: `Actions | ${app.name} | Nuon`,
  }
}

export default async function AppActionsPage({
  params,
  searchParams,
}: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
  const [{ data: app }, { data: org }] = await Promise.all([
    getAppById({ appId, orgId }),
    getOrgById({ orgId }),
  ])

  return org?.features?.['stratus-layout'] ? (
    <PageSection isScrollable>
      <HeadingGroup>
        <Text variant="base" weight="strong">
          App actions
        </Text>
      </HeadingGroup>

      {/* old layout stuff */}
      <ErrorBoundary fallback={<>Error loading app actions</>}>
        <Suspense
          fallback={<ActionsTableSkeleton />}
        >
          <ActionsTable
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
        { href: `/${orgId}/apps/${app.id}/actions`, text: 'Actions' },
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
      <Section childrenClassName="flex flex-auto">
        <OldErrorBoundary fallbackRender={ErrorFallback}>
          <Suspense
            fallback={
              <Loading variant="page" loadingText="Loading actions..." />
            }
          >
            <AppActions
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
