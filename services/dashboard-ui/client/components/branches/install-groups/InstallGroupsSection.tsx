import { EmptyState } from '@/components/common/EmptyState'
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
      <div className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg p-6">
        <EmptyState
          variant="diagram"
          emptyTitle="No install groups configured"
          emptyMessage={`Use "Deployment plan" above to add deployment groups.`}
        />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {groups.map((group, idx) => (
        <div
          key={group.id || idx}
          className="border border-cool-grey-200 dark:border-dark-grey-700 rounded-lg bg-white dark:bg-dark-grey-800 p-4 space-y-3"
        >
          <div className="flex items-center justify-between gap-3 flex-wrap">
            <Text variant="base" weight="strong">
              {group.name}
            </Text>
            <div className="flex items-center gap-2 flex-wrap justify-end">
              {(group.max_parallel || 1) > 1 && (
                <Text variant="subtext" theme="neutral">
                  Max {group.max_parallel} parallel
                </Text>
              )}
            </div>
          </div>

          {group.install_ids && group.install_ids.length > 0 ? (
            <div className="space-y-1.5">
              {group.install_ids.map((installId) => {
                const install = installsById[installId]
                return (
                  <div
                    key={installId}
                    className="flex items-center justify-between gap-4 px-3 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700"
                  >
                    <div className="min-w-0">
                      {install ? (
                        <Link
                          href={`/${orgId}/installs/${install.id}`}
                          className="truncate"
                        >
                          {install.name}
                        </Link>
                      ) : (
                        <Text
                          variant="subtext"
                          theme="neutral"
                          family="mono"
                          className="truncate"
                        >
                          {installId}
                        </Text>
                      )}
                    </div>
                    {install && (
                      <div className="shrink-0">
                        <SimpleInstallStatuses
                          install={install}
                          isLabelHidden
                        />
                      </div>
                    )}
                  </div>
                )
              })}
            </div>
          ) : (
            <div className="px-3 py-3 rounded-md border border-dashed border-cool-grey-300 dark:border-dark-grey-600 text-center">
              <Text variant="subtext" theme="neutral">
                No installs in this group
              </Text>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}
