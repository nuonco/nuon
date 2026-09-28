import { Text } from '@/components/common/Text'
import { statusAccent } from '@/components/branches/graph/accents'
import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { cn } from '@/utils/classnames'
import type { TTrackGroup } from './RolloutTrack'

export interface IRolloutTiles {
  groups: TTrackGroup[]
  onSelectGroup: (groupId: string) => void
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
  stepStatusCategory(status) === 'pending'
    ? 'bg-black/10 dark:bg-white/15'
    : statusAccent(status).dot

const tilesFor = (group: TTrackGroup) => {
  const tiles = group.installs
    .map((install) => ({
      id: install.id,
      label: `${install.name}: ${install.status}`,
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

const GroupTiles = ({ group }: { group: TTrackGroup }) => {
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

export const RolloutTiles = ({ groups, onSelectGroup }: IRolloutTiles) => (
  <ol className="flex items-start overflow-x-auto">
    {groups.map((group, index) => {
      const tiles = tilesFor(group)
      const done = tiles.filter(
        (tile) => stepStatusCategory(tile.status) === 'success'
      ).length
      const failed = tiles.some(
        (tile) => stepStatusCategory(tile.status) === 'error'
      )
      return (
        <li
          key={group.id}
          className="flex min-w-40 max-w-64 flex-1 items-start"
        >
          {index > 0 ? (
            <span aria-hidden className="mt-2.5 w-6 shrink-0 border-t" />
          ) : null}
          <button
            type="button"
            onClick={() => onSelectGroup(group.id)}
            className="flex w-full min-w-0 flex-col gap-2 rounded-md p-1 text-left hover:bg-black/[0.03] dark:hover:bg-white/[0.03]"
          >
            <span className="flex items-baseline justify-between gap-2">
              <Text variant="subtext" weight="strong" className="truncate">
                {group.name}
              </Text>
              <Text
                variant="label"
                family="mono"
                theme={failed ? 'error' : 'neutral'}
              >
                {done}/{tiles.length}
              </Text>
            </span>
            <GroupTiles group={group} />
          </button>
        </li>
      )
    })}
  </ol>
)
