import { useState } from 'react'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel } from '@/components/surfaces/Panel'
import { cn } from '@/utils/classnames'
import { CommitSha } from './CommitSha'
import type { TLatestRollout, TRolloutInstallGroup } from './fixtures'
import { GroupLabels } from './GroupLabels'
import { InstallGroupPanelBody } from './InstallGroupPanelBody'
import { InstallStatusCounts } from './InstallStatusCounts'
import {
  InstallStatusTiles,
  installStatusClass,
  installStatusKey,
} from './InstallStatusTiles'

const InstallGroupCard = ({
  group,
  commit,
  highlighted,
}: {
  group: TRolloutInstallGroup
  commit: TLatestRollout['commit']
  highlighted: boolean
}) => {
  return (
    <Panel
      panelKey={`install-group-${group.id}`}
      size="half"
      heading={group.name}
      triggerButton={{
        variant: 'ghost',
        className: cn(
          '!h-auto !w-full !whitespace-normal !items-stretch flex-col !gap-3 !p-4 text-left !rounded-md !shadow-sm !bg-white dark:!bg-dark-grey-900 !text-inherit !border-[color:var(--border-color)] outline outline-2 outline-transparent !transition-[background-color,outline-color] !duration-fast !ease-cubic hover:!bg-cool-grey-50 dark:hover:!bg-white/5',
          highlighted && '!outline-primary-600 dark:!outline-primary-400'
        ),
        children: (
          <>
            <div className="flex w-full items-center justify-between gap-3">
              <span className="flex min-w-0 items-center gap-2">
                <Text variant="subtext" theme="neutral" family="mono">
                  {group.order}
                </Text>
                <Text variant="body" weight="strong" className="truncate">
                  {group.name}
                </Text>
              </span>
              <span className="flex shrink-0 items-center gap-3">
                <CommitSha commit={commit} />
                <Status status={group.status} />
              </span>
            </div>
            <InstallStatusTiles group={group} />
            <GroupLabels group={group} max={3} />
            <span className="flex w-full flex-wrap items-center gap-x-3 gap-y-1">
              <InstallStatusCounts installs={group.installs} />
              <Text variant="subtext" theme="neutral">
                Up to {group.max_parallel ?? 1} at a time ·{' '}
                {group.auto_approve_on_policies_passing
                  ? 'Auto-approve'
                  : 'Manual approval'}
              </Text>
            </span>
          </>
        ),
      }}
    >
      <InstallGroupPanelBody group={group} commit={commit} />
    </Panel>
  )
}

const orderedGroups = (groups: TRolloutInstallGroup[]) =>
  [...groups].sort((a, b) => (a.order ?? 0) - (b.order ?? 0))

const STATUS_SLICES = [
  { status: 'success', label: 'success' },
  { status: 'error', label: 'failed' },
  { status: 'in-progress', label: 'in progress' },
  { status: 'approval-awaiting', label: 'awaiting approval' },
  { status: 'cancelled', label: 'cancelled' },
  { status: 'pending', label: 'pending' },
]

const statusSlices = (group: TRolloutInstallGroup) => {
  const counts = new Map<string, number>()
  for (const install of group.installs) {
    const key = installStatusKey(install)
    counts.set(key, (counts.get(key) ?? 0) + 1)
  }
  return STATUS_SLICES.flatMap((slice) => {
    const count = counts.get(slice.status) ?? 0
    return count ? [{ ...slice, count }] : []
  })
}

export const GroupStatusBar = ({
  groups,
  onHover = () => {},
}: {
  groups: TRolloutInstallGroup[]
  onHover?: (groupId: string, open: boolean) => void
}) => (
  <div className="flex h-3 items-stretch gap-1" aria-label="Install groups">
    {groups.map((group) => {
      const total = group.installs.length
      const slices = statusSlices(group)
      const summary = slices
        .map((slice) => `${slice.count} ${slice.label}`)
        .join(', ')
      return (
        <Tooltip
          key={group.id}
          position="top"
          className="!flex h-full min-w-0 !w-auto"
          onOpenChange={(open) => {
            if (group.id) onHover(group.id, open)
          }}
          style={{ flex: `${Math.max(total, 1)} 1 0%` }}
          tipContent={
            <span className="flex flex-col gap-1">
              <span className="flex items-center gap-2">
                <Text variant="subtext" weight="strong">
                  {group.name}
                </Text>
                <Status status={group.status} />
                <Text variant="subtext" theme="neutral">
                  {total} {total === 1 ? 'install' : 'installs'}
                </Text>
              </span>
              <span className="flex items-center gap-3">
                <InstallStatusCounts
                  installs={group.installs}
                  showTotal={false}
                />
              </span>
            </span>
          }
        >
          <span
            aria-label={`${group.name}: ${total} installs, ${summary}`}
            className="flex h-full w-full items-stretch overflow-hidden rounded-sm"
          >
            {slices.length ? (
              slices.map((slice) => (
                <span
                  key={slice.status}
                  className={cn(
                    'h-full min-w-0',
                    installStatusClass(slice.status)
                  )}
                  style={{ flex: `${slice.count} 1 0%` }}
                />
              ))
            ) : (
              <span
                className={cn('h-full w-full', installStatusClass('pending'))}
              />
            )}
          </span>
        </Tooltip>
      )
    })}
  </div>
)

export const InstallGroupCards = ({
  groups,
  commit,
}: {
  groups: TRolloutInstallGroup[]
  commit: TLatestRollout['commit']
}) => {
  const ordered = orderedGroups(groups)
  const [hoveredId, setHoveredId] = useState<string>()

  return (
    <section className="flex flex-col gap-3">
      <div className="flex flex-col gap-3">
        <SectionHeader title="Install groups" />
        <GroupStatusBar
          groups={ordered}
          onHover={(groupId, open) =>
            setHoveredId((current) =>
              open ? groupId : current === groupId ? undefined : current
            )
          }
        />
      </div>
      <ol className="flex flex-col gap-3">
        {ordered.map((group) => (
          <li key={group.id}>
            <InstallGroupCard
              group={group}
              commit={commit}
              highlighted={group.id === hoveredId}
            />
          </li>
        ))}
      </ol>
    </section>
  )
}
