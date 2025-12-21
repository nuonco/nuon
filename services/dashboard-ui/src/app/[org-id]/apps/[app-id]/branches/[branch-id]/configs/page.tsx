import type { Metadata } from 'next'
import { Suspense } from 'react'
import { notFound } from 'next/navigation'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppBranch, getBranchConfigs, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { ConfigsTable, ConfigsTableSkeleton } from './configs-table'

type TBranchConfigsPageProps = TPageProps<'org-id' | 'app-id' | 'branch-id'>

export async function generateMetadata({
  params,
}: TBranchConfigsPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params
  const { data: branch } = await getAppBranch({ appId, branchId, orgId })

  return {
    title: `Configurations | ${branch?.name || 'Branch'} | Nuon`,
  }
}

export default async function BranchConfigsPage({
  params,
  searchParams,
}: TBranchConfigsPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params
  const sp = await searchParams

  const [
    { data: branch, error: branchError },
    { data: app },
    { data: org },
  ] = await Promise.all([
    getAppBranch({ appId, branchId, orgId }),
    getApp({ appId, orgId }),
    getOrg({ orgId }),
  ])

  if (branchError || !branch) {
    notFound()
  }

  return (
    <PageSection isScrollable>
      <Breadcrumbs
        breadcrumbs={[
          {
            path: `/${orgId}`,
            text: org?.name || '',
          },
          {
            path: `/${orgId}/apps`,
            text: 'Apps',
          },
          {
            path: `/${orgId}/apps/${appId}`,
            text: app?.name || '',
          },
          {
            path: `/${orgId}/apps/${appId}/branches`,
            text: 'Branches',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}`,
            text: branch?.name || '',
          },
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}/configs`,
            text: 'Configurations',
          },
        ]}
      />

      <HeadingGroup>
        <Text variant="h3" weight="stronger">
          Configuration History
        </Text>
        <Text variant="subtext" theme="info">
          All configuration versions for {branch.name}
        </Text>
      </HeadingGroup>

      <ErrorBoundary fallback={<>Error loading configurations</>}>
        <Suspense fallback={<ConfigsTableSkeleton />}>
          <ConfigsTable
            appId={appId}
            branchId={branchId}
            orgId={orgId}
            offset={sp['offset'] || '0'}
          />
        </Suspense>
      </ErrorBoundary>
    </PageSection>
  )
}
