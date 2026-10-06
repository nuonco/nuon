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
            <GroupLabels group={group} max={3} />
            <span className="flex w-full flex-wrap items-center gap-x-3 gap-y-1">
              <InstallStatusCounts group={group} />
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

const rolledOut = (group: TRolloutInstallGroup) =>
  group.installs.filter((install) => install.status === 'success').length

const GroupProgress = ({
  groups,
  onHover,
}: {
  groups: TRolloutInstallGroup[]
  onHover: (groupId: string, open: boolean) => void
}) => (
  <div
    className="flex h-2 items-center gap-1"
    aria-label="Install group progress"
  >
    {groups.map((group) => {
      const total = group.installs.length
      const done = rolledOut(group)
      return (
        <Tooltip
          key={group.id}
          position="top"
          className="!w-auto min-w-0"
          onOpenChange={(open) => {
            if (group.id) onHover(group.id, open)
          }}
          style={{ flexGrow: Math.max(total, 1) }}
          tipContentClassName="whitespace-normal"
          tipContent={
            <span className="flex flex-col gap-1">
              <Text variant="subtext" weight="strong">
                {group.name}
              </Text>
              <span className="flex items-center gap-2">
                <Status status={group.status} />
                <Text variant="subtext" theme="neutral">
                  {done} of {total}
                </Text>
              </span>
            </span>
          }
        >
          <span
            aria-label={`${group.name}: ${group.status}, ${done} of ${total}`}
            className="block h-2 overflow-hidden rounded-sm bg-black/10 transition-[height] duration-fast ease-cubic hover:h-3.5 dark:bg-white/15"
          >
            <span
              className="block h-full bg-green-600 dark:bg-green-500"
              style={{ width: total ? `${(done / total) * 100}%` : 0 }}
            />
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
      <SectionHeader title="Install groups" />
      <GroupProgress
        groups={ordered}
        onHover={(groupId, open) =>
          setHoveredId((current) =>
            open ? groupId : current === groupId ? undefined : current
          )
        }
      />
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
