import type { Metadata } from 'next'
import { Suspense } from 'react'
import { notFound } from 'next/navigation'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { getApp, getAppBranch, getOrg } from '@/lib'
import type { TPageProps } from '@/types'
import { BranchWorkflows } from './branch-workflows'

// NOTE: old layout stuff
import { ErrorBoundary as OldErrorBoundary } from 'react-error-boundary'
import {
  AppPageSubNav,
  DashboardContent,
  ErrorFallback,
  Loading,
  Section,
} from '@/components'

type TBranchPageProps = TPageProps<'org-id' | 'app-id' | 'branch-id'>

export async function generateMetadata({
  params,
}: TBranchPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params
  const { data: branch } = await getAppBranch({ appId, branchId, orgId })

  return {
    title: `${branch.name} | Branches | Nuon`,
  }
}

export default async function AppBranchDetailPage({
  params,
}: TBranchPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params

  const [{ data: branch, error }, { data: app }, { data: org }] =
    await Promise.all([
      getAppBranch({ appId, branchId, orgId }),
      getApp({ appId, orgId }),
      getOrg({ orgId }),
    ])

  if (error || !branch) {
    notFound()
  }

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
          {
            path: `/${orgId}/apps/${appId}/branches/${branchId}`,
            text: branch?.name,
          },
        ]}
      />

      <HeadingGroup>
        <Text variant="h3" weight="stronger">
          {branch.name}
        </Text>
        <ID>{branch.id}</ID>
        <Text variant="subtext" theme="info">
          Created{' '}
          <Time
            variant="subtext"
            time={branch?.created_at}
            format="relative"
          />
        </Text>
      </HeadingGroup>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mt-6">
        <LabeledValue label="Last Synced Commit">
          <Text variant="base">
            {branch.last_synced_commit ? (
              <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                {branch.last_synced_commit.slice(0, 7)}
              </code>
            ) : (
              <span className="text-gray-400">Not synced yet</span>
            )}
          </Text>
        </LabeledValue>

        <LabeledValue label="VCS Connection">
          <Text variant="base">
            {branch.connected_github_vcs_config_id ? (
              <code className="text-xs">
                {branch.connected_github_vcs_config_id.slice(0, 8)}...
              </code>
            ) : (
              <span className="text-gray-400">-</span>
            )}
          </Text>
        </LabeledValue>

        <LabeledValue label="Workflows">
          <Text variant="base">
            {branch.workflows?.length || 0} workflow
            {branch.workflows?.length !== 1 ? 's' : ''}
          </Text>
        </LabeledValue>
      </div>

      <div className="mt-8">
        <HeadingGroup>
          <Text variant="base" weight="strong">
            Workflows
          </Text>
        </HeadingGroup>

        <ErrorBoundary fallback={<>Error loading workflows</>}>
          <Suspense fallback={<div>Loading workflows...</div>}>
            <BranchWorkflows branch={branch} />
          </Suspense>
        </ErrorBoundary>
      </div>
    </PageSection>
  ) : (
    <DashboardContent
      breadcrumb={[
        { href: `/${orgId}/apps`, text: 'Apps' },
        { href: `/${orgId}/apps/${app.id}`, text: app.name },
        { href: `/${orgId}/apps/${app.id}/branches`, text: 'Branches' },
        { href: `/${orgId}/apps/${app.id}/branches/${branch.id}`, text: branch.name },
      ]}
      heading={branch.name}
      headingUnderline={branch.id}
      headingMeta={
        <>
          Created <Time time={branch?.created_at} format="relative" />
        </>
      }
      meta={<AppPageSubNav appId={appId} orgId={orgId} />}
    >
      <Section childrenClassName="flex flex-col gap-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <LabeledValue label="Last Synced Commit">
            <Text variant="base">
              {branch.last_synced_commit ? (
                <code className="text-xs bg-gray-100 px-2 py-1 rounded">
                  {branch.last_synced_commit.slice(0, 7)}
                </code>
              ) : (
                <span className="text-gray-400">Not synced yet</span>
              )}
            </Text>
          </LabeledValue>

          <LabeledValue label="VCS Connection">
            <Text variant="base">
              {branch.connected_github_vcs_config_id ? (
                <code className="text-xs">
                  {branch.connected_github_vcs_config_id.slice(0, 8)}...
                </code>
              ) : (
                <span className="text-gray-400">-</span>
              )}
            </Text>
          </LabeledValue>

          <LabeledValue label="Workflows">
            <Text variant="base">
              {branch.workflows?.length || 0} workflow
              {branch.workflows?.length !== 1 ? 's' : ''}
            </Text>
          </LabeledValue>
        </div>

        <div>
          <Text variant="h4" weight="strong" className="mb-4">
            Workflows
          </Text>
          <OldErrorBoundary fallbackRender={ErrorFallback}>
            <Suspense
              fallback={<Loading loadingText="Loading workflows..." />}
            >
              <BranchWorkflows branch={branch} />
            </Suspense>
          </OldErrorBoundary>
        </div>
      </Section>
    </DashboardContent>
  )
}
