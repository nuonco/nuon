import { useMemo } from 'react'
import { Banner } from '@/components/common/Banner'
import { EmptyState } from '@/components/common/EmptyState'
import { LabelBadge } from '@/components/common/LabelBadge'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { InstallStatuses } from '@/components/installs/InstallStatuses'
import { resolveInstallGroupMembership } from '@/components/branches/install-group-membership'
import type { TAppBranchConfig, TInstall } from '@/types'

interface IInstallGroupsSection {
  config: TAppBranchConfig
  installsById: Record<string, TInstall>
  orgId: string
  labelColors?: Record<string, string>
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
        <InstallStatuses
          install={install}
          isLabelHidden
          lazyComponents
          tooltipPosition="top"
        />
      </div>
    )}
  </div>
)

const EmptyGroupHint = ({ children }: { children: string }) => (
  <div className="px-3 py-3 rounded-md border border-dashed text-center">
    <Text variant="subtext" theme="neutral">
      {children}
    </Text>
  </div>
)

export const InstallGroupsSection = ({
  config,
  installsById,
  orgId,
  labelColors,
}: IInstallGroupsSection) => {
  const groups = config.install_groups ?? []

  const membership = useMemo(
    () => resolveInstallGroupMembership(Object.values(installsById), groups),
    [groups, installsById]
  )
  const overlappingInstallIds = new Set(
    membership.overlappingInstalls.map((install) => install.id)
  )

  if (groups.length === 0) {
    return (
      <div className="border rounded-lg p-6">
        <EmptyState
          variant="diagram"
          emptyTitle="No install groups"
          emptyMessage="This deployment plan doesn't have any install groups yet."
        />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      {groups.map((group, idx) => {
        const labelEntries = Object.entries(
          group.label_selector?.match_labels ?? {}
        )
        const isDefault = !!group.default
        const matched = membership.installsByGroup[idx]
        const overlappingInstalls = matched.filter((install) =>
          overlappingInstallIds.has(install.id)
        )

        return (
          <div
            key={group.id || idx}
            className="border rounded-lg bg-white dark:bg-dark-grey-800 p-4 flex flex-col gap-3"
          >
            <div className="flex items-center justify-between gap-3 flex-wrap">
              <div className="flex items-center gap-2 flex-wrap min-w-0">
                <Text variant="base" weight="strong">
                  {group.name}
                </Text>
                {isDefault && (
                  <Text variant="subtext" theme="neutral">
                    Default
                  </Text>
                )}
                {labelEntries.map(([k, v]) => (
                  <LabelBadge
                    key={k}
                    labelKey={k}
                    labelValue={v}
                    size="sm"
                    customColor={labelColors?.[k]}
                  />
                ))}
              </div>
              {(group.max_parallel || 1) > 1 && (
                <Text variant="subtext" theme="neutral">
                  Max {group.max_parallel} parallel
                </Text>
              )}
            </div>

            {overlappingInstalls.length > 0 && (
              <Banner theme="warn">
                <strong>Warning:</strong>{' '}
                {overlappingInstalls.length === 1
                  ? `${overlappingInstalls[0].name} also matches another group`
                  : `${overlappingInstalls.length} installs also match other groups`}
                . Each install can match only one group; update the deployment
                plan or install labels before running this branch.
              </Banner>
            )}

            {matched.length > 0 ? (
              <div className="flex flex-col gap-1.5">
                {matched.map((install) => (
                  <InstallListRow
                    key={install.id}
                    install={install}
                    installId={install.id ?? ''}
                    orgId={orgId}
                  />
                ))}
              </div>
            ) : (
              <EmptyGroupHint>
                {isDefault
                  ? 'No installs yet — every install on this branch joins this group'
                  : "No installs currently match this group's labels"}
              </EmptyGroupHint>
            )}
          </div>
        )
      })}

      {membership.unassignedInstalls.length > 0 && (
        <div className="border rounded-lg bg-white dark:bg-dark-grey-800 p-4 flex flex-col gap-3 border-dashed">
          <div className="flex items-baseline gap-2">
            <Text variant="base" weight="strong">
              Orphaned
            </Text>
            <Text variant="subtext" theme="neutral">
              — {membership.unassignedInstalls.length} install
              {membership.unassignedInstalls.length !== 1 ? 's' : ''} won&apos;t
              receive updates
            </Text>
          </div>
          <Text variant="subtext" theme="neutral">
            These installs belong to this branch but don&apos;t match any group.
            They will not be updated when this branch runs. Assign them to a
            group to include them in deployments.
          </Text>
          <div className="flex flex-col gap-1.5">
            {membership.unassignedInstalls.map((install) => (
              <InstallListRow
                key={install.id}
                install={install}
                installId={install.id}
                orgId={orgId}
              />
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
