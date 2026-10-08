import type { ReactNode } from 'react'
import { Link } from '@/components/common/Link'
import { Text } from '@/components/common/Text'
import { statusAccent } from '@/components/branches/graph/accents'
import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { cn } from '@/utils/classnames'
import type { TTrackGroup } from './RolloutTrack'

export interface IGroupPlanApproval {
  groupName: string
  banner: ReactNode
}

export interface IRolloutGroupsCard {
  groups: TTrackGroup[]
  groupHref: (groupId: string) => string
  selectedGroupId?: string
  onSelectGroup?: (groupId: string) => void
  approvals?: IGroupPlanApproval[]
}

const MAX_TILES = 60

const CATEGORY_ORDER: TStepStatusCategory[] = [
  'success',
  'error',
  'active',
  'awaiting',
  'pending',
]

const CATEGORY_STATUS: Record<TStepStatusCategory, string> = {
  success: 'success',
  error: 'error',
  active: 'in-progress',
  awaiting: 'approval-awaiting',
  pending: 'pending',
}

const tileClass = (status: string) =>
  !status || status === 'pending'
    ? 'bg-black/10 dark:bg-white/15'
    : statusAccent(status).dot

export const tilesFor = (group: TTrackGroup) => {
  const tiles = group.installs
    .map((install) => ({
      id: install.id,
      label: `${install.name}: ${install.detail || install.status}`,
      status: install.status,
    }))
    .sort(
      (a, b) =>
        CATEGORY_ORDER.indexOf(stepStatusCategory(a.status)) -
        CATEGORY_ORDER.indexOf(stepStatusCategory(b.status))
    )
  const queued = Math.max((group.plannedCount ?? 0) - tiles.length, 0)
  for (let idx = 0; idx < queued; idx += 1) {
    tiles.push({ id: `queued-${idx}`, label: 'Queued', status: 'pending' })
  }
  return tiles
}

export const GroupTiles = ({ group }: { group: TTrackGroup }) => {
  const tiles = tilesFor(group)
  if (tiles.length === 0) {
    return (
      <Text variant="subtext" theme="neutral">
        No installs
      </Text>
    )
  }

  if (tiles.length > MAX_TILES) {
    const counts = CATEGORY_ORDER.map((category) => ({
      category,
      count: tiles.filter(
        (tile) => stepStatusCategory(tile.status) === category
      ).length,
    }))
    return (
      <span className="flex h-3 w-full overflow-hidden rounded-sm bg-black/10 dark:bg-white/15">
        {counts.map(({ category, count }) =>
          count && category !== 'pending' ? (
            <span
              key={category}
              className={statusAccent(CATEGORY_STATUS[category]).dot}
              style={{ width: `${(count / tiles.length) * 100}%` }}
            />
          ) : null
        )}
      </span>
    )
  }

  return (
    <span className="grid grid-cols-10 gap-1">
      {tiles.map((tile) => (
        <span
          key={tile.id}
          title={tile.label}
          aria-label={tile.label}
          className={cn('aspect-square rounded-sm', tileClass(tile.status))}
        />
      ))}
    </span>
  )
}

const approvalFor = (approvals: IGroupPlanApproval[] | undefined, name: string) =>
  approvals?.find(
    (item) => item.groupName.toLowerCase() === name.toLowerCase()
  )

export const RolloutGroupsCard = ({
  groups,
  groupHref,
  selectedGroupId,
  onSelectGroup,
  approvals,
}: IRolloutGroupsCard) => (
  <ol className="-m-1 flex items-start overflow-x-auto p-1">
    {groups.map((group, index) => {
      const tiles = tilesFor(group)
      const done = tiles.filter(
        (tile) => stepStatusCategory(tile.status) === 'success'
      ).length
      const failed = tiles.some(
        (tile) => stepStatusCategory(tile.status) === 'error'
      )
      const complete = tiles.length > 0 && done === tiles.length
      const isSelected = group.id === selectedGroupId
      const approval = approvalFor(approvals, group.name)
      return (
        <li
          key={group.id}
          className={cn(
            'flex min-w-40 flex-1 items-start',
            approval ? 'max-w-sm' : 'max-w-64'
          )}
        >
          {index > 0 ? (
            <span aria-hidden className="mt-5 w-6 shrink-0 border-t" />
          ) : null}
          <div className="flex min-w-0 flex-1 flex-col gap-2">
            <Link
              href={groupHref(group.id)}
              aria-current={isSelected ? 'page' : undefined}
              onClick={
                onSelectGroup
                  ? (event) => {
                      event.preventDefault()
                      onSelectGroup(group.id)
                    }
                  : undefined
              }
              className={cn(
                '!flex !h-auto !w-full min-w-0 !flex-col !items-stretch !gap-2 !rounded-xl !border !bg-white !p-3 text-left !font-normal !whitespace-normal !text-inherit !no-underline shadow-sm dark:!bg-dark-grey-900',
                'hover:!border-cool-grey-400 hover:dark:!border-cool-grey-500',
                isSelected &&
                  '!border-primary-400 !bg-primary-50/60 dark:!border-primary-500 dark:!bg-primary-600/15'
              )}
            >
              <span className="flex items-baseline justify-between gap-2">
                <Text variant="subtext" weight="strong" className="truncate">
                  {group.name}
                </Text>
                <Text
                  variant="label"
                  family="mono"
                  theme={failed ? 'error' : complete ? 'success' : 'neutral'}
                >
                  {done}/{tiles.length}
                </Text>
              </span>
              <GroupTiles group={group} />
            </Link>
            {approval?.banner}
          </div>
        </li>
      )
    })}
  </ol>
)
