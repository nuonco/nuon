import { useState, type ReactNode } from 'react'
import { Duration } from '@/components/common/Duration'
import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { statusAccent } from '@/components/branches/graph/accents'
import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { cn } from '@/utils/classnames'

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
  workflowHref?: string
}

export type TTrackGroup = {
  id: string
  name: string
  status: string
  installs: TTrackInstall[]
  plannedCount?: number
  rules?: ReactNode
}

export interface IRolloutTrack {
  groups: TTrackGroup[]
  initialGroupId?: string
  emptyMessage?: string
}

const MAX_SEGMENTS = 24

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

const isReached = (status: string) =>
  !isSkipped(status) && stepStatusCategory(status) !== 'pending'

const STATUS_FOR: Record<TStepStatusCategory, string> = {
  success: 'success',
  error: 'error',
  active: 'in-progress',
  awaiting: 'approval-awaiting',
  pending: 'pending',
}

const Segments = ({ group }: { group: TTrackGroup }) => {
  const counts = tally(group)
  if (counts.total === 0) return null

  if (counts.total > MAX_SEGMENTS) {
    return (
      <span className="flex h-1.5 w-full overflow-hidden rounded-full bg-black/5 dark:bg-white/10">
        {INSTALL_ORDER.map((category) =>
          counts[category] && category !== 'pending' ? (
            <span
              key={category}
              className={statusAccent(STATUS_FOR[category]).dot}
              style={{ width: `${(counts[category] / counts.total) * 100}%` }}
            />
          ) : null
        )}
      </span>
    )
  }

  const statuses = [
    ...group.installs.map((install) => install.status),
    ...Array.from(
      { length: counts.total - group.installs.length },
      () => 'pending'
    ),
  ]
  return (
    <span className="flex h-1.5 w-full gap-0.5">
      {statuses.map((status, idx) => (
        <span
          key={idx}
          className={cn(
            'flex-1 rounded-full',
            stepStatusCategory(status) === 'pending'
              ? 'bg-black/10 dark:bg-white/15'
              : statusAccent(status).dot
          )}
        />
      ))}
    </span>
  )
}

const Step = ({
  group,
  index,
  isSelected,
  onSelect,
}: {
  group: TTrackGroup
  index: number
  isSelected: boolean
  onSelect: () => void
}) => {
  const category = stepStatusCategory(group.status)
  return (
    <button
      type="button"
      onClick={onSelect}
      aria-pressed={isSelected}
      className={cn(
        'flex w-full min-w-0 flex-col gap-2 rounded-lg border p-3 text-left transition-colors duration-fast',
        isSelected
          ? 'bg-black/5 dark:bg-white/5'
          : 'hover:bg-black/[0.03] dark:hover:bg-white/[0.03]'
      )}
    >
      <span className="flex items-center gap-2 min-w-0">
        <Text variant="label" family="mono" theme="neutral">
          {index + 1}
        </Text>
        <Text variant="body" weight="strong" className="truncate">
          {group.name}
        </Text>
        <span className="ml-auto shrink-0">
          <Status status={group.status} isWithoutText />
        </span>
      </span>
      <Segments group={group} />
      <Text
        variant="subtext"
        theme={category === 'error' ? 'error' : 'neutral'}
        className="truncate"
      >
        {groupCaption(group)}
      </Text>
    </button>
  )
}

const StatusLine = ({
  label,
  value,
}: {
  label: string
  value?: TTrackStatus
}) => (
  <span className="flex min-w-0 items-center gap-3">
    <Text variant="label" theme="neutral" className="w-24 shrink-0">
      {label}
    </Text>
    <Status status={value?.status || 'unknown'} />
    {value?.detail ? (
      <Text variant="subtext" theme="neutral" className="truncate">
        {value.detail}
      </Text>
    ) : null}
  </span>
)

const InstallDetail = ({ install }: { install: TTrackInstall }) => (
  <div className="flex flex-col gap-3 border-t bg-black/[0.02] px-3 py-3 dark:bg-white/[0.03]">
    <StatusLine label="Resources" value={install.resources} />
    <StatusLine label="Deployment" value={install.deployment} />
    <StatusLine label="Health" value={install.health} />
    <span className="flex items-center gap-4">
      {install.workflowHref ? (
        <Link href={install.workflowHref}>View workflow</Link>
      ) : null}
      {install.overviewHref ? (
        <Link href={install.overviewHref}>Install overview</Link>
      ) : null}
    </span>
  </div>
)

const InstallList = ({ group }: { group: TTrackGroup }) => {
  const [openId, setOpenId] = useState<string>()
  const installs = [...group.installs].sort(
    (a, b) =>
      INSTALL_ORDER.indexOf(stepStatusCategory(a.status)) -
      INSTALL_ORDER.indexOf(stepStatusCategory(b.status))
  )

  return (
    <div className="flex flex-col gap-3">
      <span className="flex flex-wrap items-center gap-x-4 gap-y-1">
        <Text variant="base" weight="strong">
          {group.name}
        </Text>
        <Status status={group.status} />
        {group.rules ? (
          <Text variant="subtext" theme="neutral">
            {group.rules}
          </Text>
        ) : null}
      </span>
      {installs.length === 0 ? (
        <Text variant="subtext" theme="neutral">
          {group.plannedCount
            ? `${pluralize(group.plannedCount, 'install')} will update when this group starts.`
            : 'No installs match this group.'}
        </Text>
      ) : (
        <ul className="flex flex-col border-y">
          {installs.map((install) => {
            const isOpen = openId === install.id
            return (
              <li key={install.id} className="border-t first:border-t-0">
                <button
                  type="button"
                  aria-expanded={isOpen}
                  onClick={() =>
                    setOpenId((current) =>
                      current === install.id ? undefined : install.id
                    )
                  }
                  className="grid w-full grid-cols-[minmax(0,12rem)_minmax(0,1fr)_9rem_5rem] items-center gap-4 py-2 text-left hover:bg-black/[0.03] dark:hover:bg-white/[0.03]"
                >
                  <Text variant="subtext" family="mono" className="truncate">
                    {install.name}
                  </Text>
                  <Text variant="subtext" theme="neutral" className="truncate">
                    {install.detail}
                  </Text>
                  <Status status={install.status} />
                  <span className="text-right">
                    {install.durationNs ? (
                      <Duration
                        nanoseconds={install.durationNs}
                        variant="subtext"
                        theme="neutral"
                      />
                    ) : null}
                  </span>
                </button>
                {isOpen ? <InstallDetail install={install} /> : null}
              </li>
            )
          })}
        </ul>
      )}
    </div>
  )
}

export const RolloutTrack = ({
  groups,
  initialGroupId,
  emptyMessage,
}: IRolloutTrack) => {
  const [selectedId, setSelectedId] = useState(initialGroupId)
  const fallbackId = defaultGroupId(groups)
  const selected =
    groups.find((group) => group.id === selectedId) ??
    groups.find((group) => group.id === fallbackId)

  if (groups.length === 0) {
    return emptyMessage ? (
      <Text variant="subtext" theme="neutral">
        {emptyMessage}
      </Text>
    ) : null
  }

  return (
    <div className="flex flex-col gap-6">
      <ol className="flex items-stretch overflow-x-auto">
        {groups.map((group, index) => {
          const next = groups[index + 1]
          return (
            <li key={group.id} className="flex min-w-48 flex-1 items-center">
              <Step
                group={group}
                index={index}
                isSelected={selected?.id === group.id}
                onSelect={() => setSelectedId(group.id)}
              />
              {next ? (
                <span
                  aria-hidden
                  className={cn(
                    'w-6 shrink-0 border-t',
                    !isReached(next.status) && 'border-dashed'
                  )}
                />
              ) : null}
            </li>
          )
        })}
      </ol>
      {selected ? <InstallList group={selected} /> : null}
    </div>
  )
}
