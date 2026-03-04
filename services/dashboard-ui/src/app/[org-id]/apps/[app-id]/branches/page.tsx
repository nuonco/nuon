import type { Metadata } from 'next'
import { Suspense } from 'react'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { BranchesTable, BranchesTableSkeleton } from './branches-table'
import { CreateBranchButton } from './create-branch-button'

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
import { AppBranches } from './branches'

type TAppPageProps = TPageProps<'org-id' | 'app-id'>

export async function generateMetadata({
  params,
}: TAppPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const { data: app } = await getApp({ appId, orgId })

  return {
    title: `Branches | ${app.name} | Nuon`,
  }
}

export default async function AppBranchesPage({
  params,
  searchParams,
}: TAppPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId } = await params
  const sp = await searchParams
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
            path: `/${orgId}/apps/${appId}/branches`,
            text: 'Branches',
          },
        ]}
      />
      <div className="flex items-center gap-4 justify-between">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            App branches
          </Text>
        </HeadingGroup>

        <CreateBranchButton appId={appId} orgId={orgId} />
      </div>

      <ErrorBoundary fallback={<>Error loading branches</>}>
        <Suspense fallback={<BranchesTableSkeleton />}>
          <BranchesTable
            appId={appId}
            orgId={orgId}
            offset={sp['offset'] || '0'}
          />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
        { href: `/${orgId}/apps/${app.id}/branches`, text: 'Branches' },
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
              <Loading variant="page" loadingText="Loading branches..." />
            }
          >
            <AppBranches appId={appId} orgId={orgId} offset={sp['offset'] || '0'} />
          </Suspense>
        </OldErrorBoundary>
      </Section>
    </DashboardContent>
  )
}