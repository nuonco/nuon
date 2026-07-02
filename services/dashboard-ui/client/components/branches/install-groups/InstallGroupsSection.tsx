import { EmptyState } from '@/components/common/EmptyState'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { matchesSelector } from '@/components/match/matches'
import type { TAppBranchConfig, TInstall } from '@/types'

interface IInstallGroupsSection {
  config: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
}

const InstallListRow = ({
  install,
  installId,
  orgId,
}: {
  install?: TInstall
  installId: string
  orgId: string
}) => (
  <div className="flex items-center justify-between gap-4 px-3 py-2 rounded-md bg-cool-grey-50 dark:bg-dark-grey-700">
    <div className="min-w-0">
      {install ? (
        <Link href={`/${orgId}/installs/${install.id}`} className="truncate">
          {install.name}
        </Link>
      ) : (
        <Text variant="subtext" theme="neutral" family="mono" className="truncate">
          {installId}
        </Text>
      )}
    </div>
    {install && (
      <div className="shrink-0">
        <InstallStatuses install={install} isLabelHidden lazyComponents tooltipPosition="top" />
      </div>
    )}
  </div>
)

const EmptyGroupHint = ({ children }: { children: string }) => (
  <div className="px-3 py-3 rounded-md border border-dashed text-center">
    <Text variant="subtext" theme="neutral">{children}</Text>
  </div>
)

export const InstallGroupsSection = ({
  config,
  installsById,
  orgId,
}: IInstallGroupsSection) => {
  const groups = config.install_groups ?? []

  if (groups.length === 0) {
    return (
      <div className="border rounded-lg p-6">
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
      {groups.map((group, idx) => {
        const labelEntries = Object.entries(group.label_selector?.match_labels ?? {})
        const isLabels = labelEntries.length > 0
        const matched = isLabels
          ? Object.values(installsById).filter((i) => matchesSelector(i.labels, group.label_selector))
          : []
        const installIds = group.install_ids ?? []

        return (
          <div
            key={group.id || idx}
            className="border rounded-lg bg-white dark:bg-dark-grey-800 p-4 flex flex-col gap-3"
          >
            <div className="flex items-center justify-between gap-3 flex-wrap">
              <div className="flex items-center gap-2 flex-wrap min-w-0">
                <Text variant="base" weight="strong">{group.name}</Text>
                {labelEntries.map(([k, v]) => (
                  <LabelBadge key={k} labelKey={k} labelValue={v} size="sm" variant="code" />
                ))}
              </div>
              {(group.max_parallel || 1) > 1 && (
                <Text variant="subtext" theme="neutral">
                  Max {group.max_parallel} parallel
                </Text>
              )}
            </div>

            {isLabels ? (
              matched.length > 0 ? (
                <div className="flex flex-col gap-1.5">
                  {matched.map((install) => (
                    <InstallListRow key={install.id} install={install} installId={install.id ?? ''} orgId={orgId} />
                  ))}
                </div>
              ) : (
                <EmptyGroupHint>No installs currently match this group&apos;s labels</EmptyGroupHint>
              )
            ) : installIds.length > 0 ? (
              <div className="flex flex-col gap-1.5">
                {installIds.map((installId) => (
                  <InstallListRow
                    key={installId}
                    install={installsById[installId]}
                    installId={installId}
                    orgId={orgId}
                  />
                ))}
              </div>
            ) : (
              <EmptyGroupHint>No installs in this group</EmptyGroupHint>
            )}
          </div>
        )
      })}
    </div>
  )
}
