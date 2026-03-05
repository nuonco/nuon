import type { Metadata } from 'next'
import { notFound } from 'next/navigation'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Text } from '@/components/common/Text'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppBranch, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { RunsTable } from './runs-table'

type TBranchRunsPageProps = TPageProps<'org-id' | 'app-id' | 'branch-id'>

export async function generateMetadata({
  params,
}: TBranchRunsPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params
  const { data: branch } = await getAppBranch({ appId, branchId, orgId })

  return {
    title: `Workflow Runs | ${branch?.name || 'Branch'} | Nuon`,
  }
}

export default async function BranchRunsPage({
  params,
}: TBranchRunsPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params

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
            path: `/${orgId}/apps/${appId}/branches/${branchId}/runs`,
            text: 'Workflow Runs',
          },
        ]}
      />

      <HeadingGroup>
        <Text variant="h3" weight="stronger">
          Workflow Runs
        </Text>
        <Text variant="subtext" theme="info">
          All workflow runs for {branch.name}
        </Text>
      </HeadingGroup>

      <RunsTable appId={appId} branchId={branchId} orgId={orgId} />
    </PageSection>
  )
}
