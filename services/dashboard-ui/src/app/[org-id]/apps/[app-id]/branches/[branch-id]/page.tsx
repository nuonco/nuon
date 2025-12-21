import type { Metadata } from 'next'
import { Suspense } from 'react'
import { notFound } from 'next/navigation'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { ID } from '@/components/common/ID'
import { Text } from '@/components/common/Text'
import { Link } from '@/components/common/Link'
import { Icon } from '@/components/common/Icon'
import { LabeledValue } from '@/components/common/LabeledValue'
import { Time } from '@/components/common/Time'
import { Card } from '@/components/common/Card'
import { Badge } from '@/components/common/Badge'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { TemporalLink } from '@/components/admin/TemporalLink'
import { getApp, getAppBranch, getOrg, listQueues } from '@/lib'
import type { TPageProps } from '@/types'
import { BranchDetailActions } from './branch-detail-actions'
import { BranchWorkflowRunsTable } from './branch-workflow-runs-table'

type TBranchPageProps = TPageProps<'org-id' | 'app-id' | 'branch-id'>

export async function generateMetadata({
  params,
}: TBranchPageProps): Promise<Metadata> {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params
  const { data: branch } = await getAppBranch({ appId, branchId, orgId })

  return {
    title: `${branch?.name || 'Branch'} | Branches | Nuon`,
  }
}

export default async function AppBranchDetailPage({
  params,
}: TBranchPageProps) {
  const { ['org-id']: orgId, ['app-id']: appId, ['branch-id']: branchId } =
    await params

  const [
    { data: branch, error },
    { data: app },
    { data: org },
    { data: queues },
  ] = await Promise.all([
    getAppBranch({ appId, branchId, orgId, latestConfig: true }),
    getApp({ appId, orgId }),
    getOrg({ orgId }),
    listQueues({ orgId, ownerId: branchId, ownerType: 'app_branches' }),
  ])

  // Debug logging
  if (error) {
    console.error('Error fetching branch:', error)
    console.error('Branch ID:', branchId)
    console.error('App ID:', appId)
    console.error('Org ID:', orgId)
  }

  if (error || !branch) {
    notFound()
  }

  // Get current config (most recent)
  const currentConfig =
    branch.configs && branch.configs.length > 0
      ? branch.configs.sort(
          (a, b) => (b.config_number || 0) - (a.config_number || 0)
        )[0]
      : undefined

  // Get the branch's queue for Temporal link
  const branchQueue = queues && queues.length > 0 ? queues[0] : undefined

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
        ]}
      />

      {/* Page Header */}
      <div className="flex items-start justify-between mb-6">
        <HeadingGroup>
          <Text variant="h3" weight="stronger">
            {branch.name}
          </Text>
          <ID>{branch.id}</ID>
          <Text variant="subtext" theme="info">
            Created{' '}
            <Time time={branch?.created_at} format="relative" />
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
          {branchQueue && (
            <TemporalLink 
              namespace="apps" 
              eventLoopId={`queue-${branchQueue.id}`}
              skipPrefix
            />
          )}
          <BranchDetailActions
            branch={branch}
            currentConfig={currentConfig}
            appId={appId}
            orgId={orgId}
          />
        </div>
      </div>

      {/* Install Groups Section */}
      <Card className="mb-6">
        <div className="p-6">
          <div className="flex items-center justify-between mb-4">
            <Text variant="h4" weight="strong">
              Install Groups
            </Text>
            {currentConfig && (
              <Badge theme="info" size="sm">
                v{currentConfig.config_number}
              </Badge>
            )}
          </div>

          {!currentConfig ? (
            <div className="text-center py-8">
              <Text variant="body" theme="neutral">
                No configuration yet. Click &quot;Edit&quot; above to set up install groups.
              </Text>
            </div>
          ) : currentConfig.install_groups && currentConfig.install_groups.length > 0 ? (
            <div className="space-y-3">
              {currentConfig.install_groups.map((group, idx) => (
                <div
                  key={group.id || idx}
                  className="p-4 bg-gray-50 dark:bg-gray-900 rounded-md"
                >
                  <div className="flex items-center justify-between mb-2">
                    <Text variant="base" weight="strong">
                      {idx + 1}. {group.name}
                    </Text>
                    <div className="flex items-center gap-2">
                      {group.requires_approval && (
                        <Badge theme="warning" size="sm">
                          Requires Approval
                        </Badge>
                      )}
                      {group.rollback_on_failure && (
                        <Badge theme="info" size="sm">
                          Rollback on Failure
                        </Badge>
                      )}
                    </div>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-sm">
                    <div>
                      <Text variant="subtext" theme="neutral">
                        {group.install_ids?.length || 0} install{group.install_ids?.length !== 1 ? 's' : ''}
                      </Text>
                    </div>
                    <div>
                      <Text variant="subtext" theme="neutral">
                        Max {group.max_parallel || 1} parallel
                      </Text>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <Text variant="body" theme="neutral">
                No install groups configured. Click &quot;Edit&quot; above to add deployment groups.
              </Text>
            </div>
          )}
        </div>
      </Card>

      {/* Workflow Runs Section */}
      <div>
        <div className="flex items-center justify-between mb-4">
          <Text variant="h4" weight="strong">
            Workflow Runs
          </Text>
          <Link href={`/${orgId}/apps/${appId}/branches/${branchId}/runs`}>
            View All <Icon variant="CaretRightIcon" />
          </Link>
        </div>
        <ErrorBoundary fallback={<>Error loading workflow runs</>}>
          <Suspense fallback={<div>Loading workflow runs...</div>}>
            <BranchWorkflowRunsTable
              appId={appId}
              branchId={branchId}
              orgId={orgId}
            />
          </Suspense>
        </ErrorBoundary>
      </div>
    </PageSection>
  )
}