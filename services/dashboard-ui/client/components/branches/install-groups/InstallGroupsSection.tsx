import { Badge } from '@/components/common/Badge'
import { EmptyState } from '@/components/common/EmptyState'
import { ID } from '@/components/common/ID'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { SimpleInstallStatuses } from '@/components/installs/InstallStatuses'
import type { TAppBranchConfig, TInstall } from '@/types'

interface IInstallGroupsSection {
  config: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
}

export const InstallGroupsSection = ({
  config,
  installsById,
  orgId,
}: IInstallGroupsSection) => {
  const groups = config.install_groups ?? []

  if (groups.length === 0) {
    return (
      <EmptyState
        variant="diagram"
        emptyTitle="No install groups configured"
        emptyMessage={`Use "Manage installs" above to add deployment groups.`}
      />
    )
  }

  return (
    <div className="space-y-0">
      {groups.map((group, idx) => (
        <div key={group.id || idx}>
          {idx > 0 && (
            <div className="border-t border-cool-grey-200 dark:border-dark-grey-700 my-4" />
          )}
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Badge variant="code" theme="neutral" size="sm">
                  {idx + 1}
                </Badge>
                <Text variant="base" weight="strong">
                  {group.name}
                </Text>
              </div>
              <div className="flex items-center gap-2 flex-wrap justify-end">
                {group.requires_approval && (
                  <Badge theme="warn" size="sm">
                    Requires approval
                  </Badge>
                )}
                {group.rollback_on_failure && (
                  <Badge theme="info" size="sm">
                    Rollback on failure
                  </Badge>
                )}
                <Text variant="subtext" theme="neutral">
                  Max {group.max_parallel || 1} parallel
                </Text>
              </div>
            </div>

            <div className="space-y-1.5">
              {group.install_ids && group.install_ids.length > 0 ? (
                group.install_ids.map((installId) => {
                  const install = installsById[installId]
                  return (
                    <div
                      key={installId}
                      className="flex items-center justify-between gap-4 px-3 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-900"
                    >
                      <div className="flex flex-col min-w-0">
                        {install ? (
                          <Link href={`/${orgId}/installs/${install.id}`} className="truncate">
                            {install.name}
                          </Link>
                        ) : (
                          <Text variant="subtext" theme="neutral" family="mono" className="truncate">
                            {installId}
                          </Text>
                        )}
                        <ID>{installId}</ID>
                      </div>
                      {install && (
                        <div className="shrink-0">
                          <SimpleInstallStatuses install={install} isLabelHidden />
                        </div>
                      )}
                    </div>
                  )
                })
              ) : (
                <div className="px-3 py-4 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
                  <Text variant="subtext" theme="neutral">
                    No installs in this group
                  </Text>
                </div>
              )}
            </div>
          </div>
        </div>
      ))}
    </div>
  )
}
