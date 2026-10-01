import { useState, type ReactNode } from 'react'
import { Duration } from '@/components/common/Duration'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Button } from '@/components/common/Button'
import { LabelBadge } from '@/components/common/LabelBadge'
import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { InstallGroupMatch } from './InstallGroupMatch'
import { RolloutTiles } from './RolloutTiles'

export type TTrackStatus = {
  status: string
  detail?: string
}

export type TTrackInstall = {
  id: string
  name: string
  status: string
  detail?: string
  durationNs?: number
  href?: string
  resources?: TTrackStatus
  deployment?: TTrackStatus
  health?: TTrackStatus
  overviewHref?: string
  workflowId?: string
  workflowHref?: string
  labels?: Record<string, string>
  region?: string
}

export type TGroupMatch = {
  kind: 'labels' | 'default' | 'pinned'
  labels?: Record<string, string>
}

export type TTrackGroup = {
  id: string
  name: string
  status: string
  installs: TTrackInstall[]
  plannedCount?: number
  rules?: ReactNode
  match?: TGroupMatch
  approval?: string
  maxParallel?: number
}

export interface IRolloutTrack {
  groups: TTrackGroup[]
  initialGroupId?: string
  selectedGroupId?: string
  emptyMessage?: string
  summary?: ReactNode
  notice?: ReactNode
  footer?: ReactNode
  onSelectGroup?: (groupId: string) => void
  onSelectInstall?: (groupId: string, installId: string) => void
}

const INSTALL_ORDER: TStepStatusCategory[] = [
  'error',
  'active',
  'awaiting',
  'success',
  'pending',
]

const isSkipped = (status: string) => status.endsWith('skipped')

const pluralize = (count: number, noun: string) =>
  `${count} ${noun}${count === 1 ? '' : 's'}`

const tally = (group: TTrackGroup) => {
  const counts = { success: 0, error: 0, active: 0, awaiting: 0, pending: 0 }
  group.installs.forEach((install) => {
    counts[stepStatusCategory(install.status)] += 1
  })
  const total = group.installs.length || group.plannedCount || 0
  counts.pending += total - group.installs.length
  return { ...counts, total }
}

const groupCaption = (group: TTrackGroup) => {
  const { success, error, total } = tally(group)
  if (isSkipped(group.status)) return 'Skipped'
  switch (stepStatusCategory(group.status)) {
    case 'awaiting':
      return 'Waiting for approval'
    case 'error':
      return `${error} failed · ${success} of ${total} updated`
    case 'active':
      return `${success} of ${total} updated`
    case 'success':
      return `${pluralize(total, 'install')} updated`
    default:
      return total ? `${pluralize(total, 'install')} queued` : 'No installs'
  }
}

export const selectedTrackGroup = (groups: TTrackGroup[], groupId?: string) =>
  groups.find((group) => group.id === groupId) ??
  groups.find((group) => group.id === defaultGroupId(groups))

const defaultGroupId = (groups: TTrackGroup[]) => {
  const byCategory = (category: TStepStatusCategory) =>
    groups.find((group) => stepStatusCategory(group.status) === category)
  return (
    byCategory('error') ??
    byCategory('awaiting') ??
    byCategory('active') ??
    (groups.every((group) => stepStatusCategory(group.status) === 'success')
      ? groups.at(-1)
      : groups[0])
  )?.id
}

const groupPace = (group: TTrackGroup) =>
  [
    group.maxParallel != null
      ? `Up to ${group.maxParallel} at a time`
      : undefined,
    group.approval,
  ]
    .filter(Boolean)
    .join(' · ')

const GroupHeading = ({ group }: { group: TTrackGroup }) => {
  const pace = groupPace(group)
  return (
    <div className="flex flex-col gap-1">
      <span className="flex flex-wrap items-center gap-x-4 gap-y-1">
        <Text variant="base" weight="strong">
          {group.name}
        </Text>
        <Status status={group.status} />
        <Text
          variant="subtext"
          theme={
            stepStatusCategory(group.status) === 'error' ? 'error' : 'neutral'
          }
        >
          {groupCaption(group)}
        </Text>
      </span>
      {pace ? (
        <Text variant="subtext" theme="neutral">
          {pace}
        </Text>
      ) : null}
      {group.match ? (
        <InstallGroupMatch match={group.match} />
      ) : group.rules ? (
        <Text variant="subtext" theme="neutral">
          {group.rules}
        </Text>
      ) : null}
    </div>
  )
}

const InstallList = ({
  group,
  onSelectInstall,
}: {
  group: TTrackGroup
  onSelectInstall?: (groupId: string, installId: string) => void
}) => {
  const installs = [...group.installs].sort(
    (a, b) =>
      INSTALL_ORDER.indexOf(stepStatusCategory(a.status)) -
      INSTALL_ORDER.indexOf(stepStatusCategory(b.status))
  )

  return installs.length === 0 ? (
    <Text variant="subtext" theme="neutral">
      {group.plannedCount
        ? `${pluralize(group.plannedCount, 'install')} will update when this group starts.`
        : 'No installs match this group.'}
    </Text>
  ) : (
    <ul className="flex flex-col gap-2">
      {installs.map((install) => (
        <li
          key={install.id}
          className="flex items-center gap-4 rounded-xl border bg-white px-4 py-3 shadow-sm dark:bg-dark-grey-900"
        >
          <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-3 gap-y-2">
            <Text variant="subtext" family="mono" className="truncate">
              {install.name}
            </Text>
            {Object.entries(install.labels ?? {}).map(([key, value]) => (
              <LabelBadge
                key={key}
                labelKey={key}
                labelValue={value}
                size="sm"
              />
            ))}
            {install.region ? (
              <Text variant="subtext" theme="neutral">
                {install.region}
              </Text>
            ) : null}
            {install.detail ? (
              <Text variant="subtext" theme="neutral" className="truncate">
                {install.detail}
              </Text>
            ) : null}
            <Status status={install.status} />
            {install.durationNs ? (
              <Duration
                nanoseconds={install.durationNs}
                variant="subtext"
                theme="neutral"
              />
            ) : null}
          </div>
          {onSelectInstall ? (
            <Button
              size="sm"
              variant="secondary"
              onClick={() => onSelectInstall(group.id, install.id)}
            >
              Details
            </Button>
          ) : null}
        </li>
      ))}
    </ul>
  )
}

export const RolloutTrack = ({
  groups,
  initialGroupId,
  selectedGroupId,
  emptyMessage,
  summary,
  notice,
  footer,
  onSelectGroup,
  onSelectInstall,
}: IRolloutTrack) => {
  const [localId, setLocalId] = useState(initialGroupId)
  const fallbackId = defaultGroupId(groups)
  const chosenId = onSelectGroup ? selectedGroupId : localId
  const selected =
    groups.find((group) => group.id === chosenId) ??
    groups.find((group) => group.id === fallbackId)
  const choose = (id: string) => {
    if (onSelectGroup) onSelectGroup(id)
    else setLocalId(id)
  }

  if (groups.length === 0) {
    return emptyMessage ? (
      <Text variant="subtext" theme="neutral">
        {emptyMessage}
      </Text>
    ) : null
  }

  return (
    <div className="flex flex-col gap-6">
      <RolloutTiles
        groups={groups}
        selectedGroupId={selected?.id}
        onSelectGroup={choose}
      />
      {selected ? (
        <section className="flex flex-col gap-4">
          <GroupHeading group={selected} />
          {notice}
          {footer}
          {summary ?? (
            <InstallList group={selected} onSelectInstall={onSelectInstall} />
          )}
        </section>
      ) : null}
    </div>
  )
}
