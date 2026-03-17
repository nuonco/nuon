import { useParams } from 'react-router'
import { useQuery } from '@tanstack/react-query'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { PageSection } from '@/components/layout/PageSection'
import { Breadcrumbs } from '@/components/navigation/Breadcrumb'
import { PageTitle } from '@/components/navigation/PageTitle'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'

const CONTAINER_ID = 'branch-detail-page'
import { BranchDetailActions } from '@/components/branches/BranchDetailActions'
import { getBranchWorkflowRuns } from '@/lib'

const BranchDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string

  // Fetch recent runs (limit to 5 for preview)
  const { data: runs = [] } = useQuery({
    queryKey: ['branch-runs', orgId, appId, branchId],
    queryFn: () =>
      getBranchWorkflowRuns({
        orgId,
        appId,
        branchId,
        limit: 5,
      }),
    enabled: !!orgId && !!appId && !!branchId,
    refetchInterval: 5000, // Poll every 5 seconds
  })

  // Get current config (most recent)
  const currentConfig =
    branch.configs && branch.configs.length > 0
      ? branch.configs.sort(
          (a, b) => (b.config_number || 0) - (a.config_number || 0)
        )[0]
      : undefined

  return (
    <PageSection id={CONTAINER_ID} isScrollable>
      <PageTitle title={`${branch?.name ?? 'Branch'} | ${app?.name}`} />
      <Breadcrumbs
        breadcrumbs={[
          { path: `/${org?.id}`, text: org?.name },
          { path: `/${org?.id}/apps`, text: 'Apps' },
          { path: `/${org?.id}/apps/${app?.id}`, text: app?.name },
          { path: `/${org?.id}/apps/${app?.id}/branches`, text: 'Branches' },
          { path: `/${org?.id}/apps/${app?.id}/branches/${branchId}`, text: branch?.name },
        ]}
      />
      {/* Page Header */}
      <div className="flex items-start justify-between">
        <HeadingGroup>
          <Text variant="h3" weight="stronger">
            {branch.name}
          </Text>
          <ID>{branch.id}</ID>
          <Text variant="subtext" theme="info">
            Created <Time time={branch?.created_at} format="relative" />
          </Text>
        </HeadingGroup>

        <div className="flex items-center gap-4">
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
                No configuration yet. Click "Edit" above to set up install groups.
              </Text>
            </div>
          ) : currentConfig.install_groups &&
            currentConfig.install_groups.length > 0 ? (
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
                        {group.install_ids?.length || 0} install
                        {group.install_ids?.length !== 1 ? 's' : ''}
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
                No install groups configured. Click "Edit" above to add
                deployment groups.
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

        {runs.length === 0 ? (
          <div className="text-center py-8 border border-dashed border-gray-300 dark:border-gray-700 rounded-md">
            <Text variant="body" theme="neutral">
              No workflow runs yet. Click "Trigger Run" above to start a deployment.
            </Text>
          </div>
        ) : (
          <div className="space-y-2">
            {runs.map((run) => {
              const status = run.status?.status || 'unknown'
              const statusDescription = run.status?.status_human_description || ''
              
              // Get commit info from app_branch_runs if available
              const branchRun = run.app_branch_runs?.[0]
              const commitId = branchRun?.commit_sha || branchRun?.commit_id
              
              return (
                <Link
                  key={run.id}
                  href={`/${orgId}/apps/${appId}/branches/${branchId}/runs/${run.id}`}
                  className="block p-4 border border-gray-200 dark:border-gray-800 rounded-md hover:bg-gray-50 dark:hover:bg-gray-900 transition-colors"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-2">
                        <Badge
                          theme={
                            status === 'success' || status === 'completed'
                              ? 'success'
                              : status === 'failed' || status === 'error'
                              ? 'error'
                              : status === 'running' || status === 'in-progress'
                              ? 'info'
                              : 'neutral'
                          }
                          size="sm"
                        >
                          {status}
                        </Badge>
                        <Text variant="base" weight="strong" className="truncate">
                          {statusDescription || `Run #${run.id?.substring(0, 8)}`}
                        </Text>
                      </div>
                      {commitId && (
                        <div className="flex items-center gap-2">
                          <Icon variant="GitCommit" size={14} />
                          <Text variant="subtext" theme="neutral" className="font-mono text-xs">
                            {commitId.substring(0, 7)}
                          </Text>
                        </div>
                      )}
                    </div>
                    <Text variant="subtext" theme="neutral" className="flex-shrink-0">
                      <Time time={run.created_at} format="relative" />
                    </Text>
                  </div>
                </Link>
              )
            })}
          </div>
        )}
      </div>
    </PageSection>
  )
}

export const BranchDetail = () => {
  const params = useParams()
  const branchId = params.branchId as string

  return (
    <BranchProvider branchId={branchId} shouldPoll>
      <BranchDetailContent />
    </BranchProvider>
  )
}