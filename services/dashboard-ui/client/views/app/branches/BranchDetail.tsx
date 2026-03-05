import { useParams } from 'react-router'
import { Badge } from '@/components/common/Badge'
import { Card } from '@/components/common/Card'
import { HeadingGroup } from '@/components/common/HeadingGroup'
import { Icon } from '@/components/common/Icon'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { Time } from '@/components/common/Time'
import { useApp } from '@/hooks/use-app'
import { useOrg } from '@/hooks/use-org'
import { useBranch } from '@/hooks/use-branch'
import { BranchProvider } from '@/providers/branch-provider'
import { BranchDetailActions } from './components/BranchDetailActions'

const BranchDetailContent = () => {
  const { org } = useOrg()
  const { app } = useApp()
  const { branch } = useBranch()
  const params = useParams()
  const orgId = params.orgId as string
  const appId = params.appId as string
  const branchId = params.branchId as string

  // Get current config (most recent)
  const currentConfig =
    branch.configs && branch.configs.length > 0
      ? branch.configs.sort(
          (a, b) => (b.config_number || 0) - (a.config_number || 0)
        )[0]
      : undefined

  return (
    <div className="flex flex-col gap-6 p-6">
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
        <Text variant="body" theme="neutral">
          Workflow runs will appear here
        </Text>
      </div>
    </div>
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