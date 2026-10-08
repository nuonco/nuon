import { useState, type ReactNode } from 'react'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { SectionHeader } from '@/components/layout/SectionHeader'
import { Panel } from '@/components/surfaces/Panel'
import { cn } from '@/utils/classnames'
import type {
  TTrackGroup,
  TTrackInstall,
} from '@/components/branches/BranchOverview/RolloutTrack'
import type { IGroupPlanApproval } from '@/components/branches/BranchOverview/RolloutGroupsCard'
import { CommitRange, type ICommitRange } from './CommitRange'
import { GroupLabels } from './GroupLabels'
import { InstallGroupPanelBody } from './InstallGroupPanelBody'
import { InstallStatusCounts } from './InstallStatusCounts'
import { InstallStatusTiles } from './InstallStatusTiles'
import {
  installStatusClass,
  statusSlices,
  withQueuedInstalls,
} from './install-status'

const pace = (group: TTrackGroup) =>
  [
    group.maxParallel != null
      ? `Up to ${group.maxParallel} at a time`
      : undefined,
    group.approval,
  ]
    .filter(Boolean)
    .join(' · ')

const InstallGroupCard = ({
  group,
  order,
  commit,
  highlighted,
  approval,
  onSelectInstall,
}: {
  group: TTrackGroup
  order: number
  commit?: ICommitRange
  highlighted: boolean
  approval?: ReactNode
  onSelectInstall?: (install: TTrackInstall) => void
}) => {
  const installs = withQueuedInstalls(group.installs, group.plannedCount)
  const cadence = pace(group)

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
                  {order}
                </Text>
                <Text variant="body" weight="strong" className="truncate">
                  {group.name}
                </Text>
              </span>
              <span className="flex shrink-0 items-center gap-3">
                {commit ? <CommitRange commit={commit} /> : null}
                <Status status={group.status} />
              </span>
            </div>
            <InstallStatusTiles installs={installs} />
            <GroupLabels match={group.match} max={3} />
            <span className="flex w-full flex-wrap items-center gap-x-3 gap-y-1">
              <InstallStatusCounts installs={installs} />
              {cadence ? (
                <Text variant="subtext" theme="neutral">
                  {cadence}
                </Text>
              ) : null}
            </span>
          </>
        ),
      }}
    >
      <InstallGroupPanelBody
        group={group}
        commit={commit}
        approval={approval}
        onSelectInstall={onSelectInstall}
      />
    </Panel>
  )
}

export const GroupStatusBar = ({
  groups,
  onHover = () => {},
}: {
  groups: TTrackGroup[]
  onHover?: (groupId: string, open: boolean) => void
}) => (
  <div className="flex h-3 items-stretch gap-1" aria-label="Install groups">
    {groups.map((group) => {
      const installs = withQueuedInstalls(group.installs, group.plannedCount)
      const total = installs.length
      const slices = statusSlices(installs)
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
                <InstallStatusCounts installs={installs} showTotal={false} />
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

const approvalFor = (
  approvals: IGroupPlanApproval[] | undefined,
  name: string
) =>
  approvals?.find((item) => item.groupName.toLowerCase() === name.toLowerCase())

export const InstallGroupCards = ({
  groups,
  commit,
  approvals,
  onSelectInstall,
}: {
  groups: TTrackGroup[]
  commit?: ICommitRange
  approvals?: IGroupPlanApproval[]
  onSelectInstall?: (install: TTrackInstall) => void
}) => {
  const [hoveredId, setHoveredId] = useState<string>()

  return (
    <section className="flex flex-col gap-3">
      <div className="flex flex-col gap-3">
        <SectionHeader title="Install groups" />
        <GroupStatusBar
          groups={groups}
          onHover={(groupId, open) =>
            setHoveredId((current) =>
              open ? groupId : current === groupId ? undefined : current
            )
          }
        />
      </div>
      <ol className="flex flex-col gap-3">
        {groups.map((group, index) => (
          <li key={group.id}>
            <InstallGroupCard
              group={group}
              order={index + 1}
              commit={commit}
              highlighted={group.id === hoveredId}
              approval={approvalFor(approvals, group.name)?.banner}
              onSelectInstall={onSelectInstall}
            />
          </li>
        ))}
      </ol>
    </section>
  )
}
