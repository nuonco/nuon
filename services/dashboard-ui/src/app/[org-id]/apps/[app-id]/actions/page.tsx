import type { Metadata } from 'next'
import { Suspense } from 'react'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { PageSection } from '@/components/layout/PageSection'
import { Text } from '@/components/common/Text'
import { getAppById, getOrgById } from '@/lib'
import type { TPageProps } from '@/types'
import { AppActions } from './actions'

// NOTE: old layout stuff
import { ErrorBoundary } from 'react-error-boundary'
import {
  AppCreateInstallButton,
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
      <ErrorBoundary fallbackRender={ErrorFallback}>
        <Suspense
          fallback={<Loading variant="page" loadingText="Loading actions..." />}
        >
          <AppActions
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
        <ErrorBoundary fallbackRender={ErrorFallback}>
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
        </ErrorBoundary>
      </Section>
    </DashboardContent>
  )
}
