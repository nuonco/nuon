import type { ReactNode } from 'react'
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
import type { TTrackGroup } from './RolloutTrack'

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

export const groupCaption = (group: TTrackGroup) => {
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

const groupPace = (group: TTrackGroup) =>
  [
    group.maxParallel != null
      ? `Up to ${group.maxParallel} at a time`
      : undefined,
    group.approval,
  ]
    .filter(Boolean)
    .join(' · ')

export const GroupHeading = ({ group }: { group: TTrackGroup }) => {
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
      {group.match ? (
        <InstallGroupMatch match={group.match} pace={pace || undefined} />
      ) : pace ? (
        <Text variant="subtext" theme="neutral">
          {pace}
        </Text>
      ) : group.rules ? (
        <Text variant="subtext" theme="neutral">
          {group.rules}
        </Text>
      ) : null}
    </div>
  )
}

export const InstallList = ({
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
              View details
            </Button>
          ) : null}
        </li>
      ))}
    </ul>
  )
}

export interface IRolloutGroupDetail {
  group: TTrackGroup
  notice?: ReactNode
  footer?: ReactNode
  summary?: ReactNode
  onSelectInstall?: (groupId: string, installId: string) => void
}

export const RolloutGroupDetail = ({
  group,
  notice,
  footer,
  summary,
  onSelectInstall,
}: IRolloutGroupDetail) => (
  <section className="flex flex-col gap-4">
    <GroupHeading group={group} />
    {notice}
    {footer}
    {summary ?? <InstallList group={group} onSelectInstall={onSelectInstall} />}
  </section>
)
