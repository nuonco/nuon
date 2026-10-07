import {
  stepStatusCategory,
  type TStepStatusCategory,
} from '@/components/branches/shared/step-status'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import type { TTheme } from '@/types'
import { cn } from '@/utils/classnames'
import { getStatusTheme } from '@/utils/status-utils'
import type { TRolloutInstall, TRolloutInstallGroup } from './fixtures'

const STATUS_FILL: Record<TTheme, string> = {
  default: 'bg-cool-grey-600 dark:bg-white/70',
  neutral: 'bg-cool-grey-600 dark:bg-white/70',
  success: 'bg-green-600 dark:bg-green-500',
  error: 'bg-red-600 dark:bg-red-500',
  warn: 'bg-orange-600 dark:bg-orange-500',
  info: 'bg-blue-600 dark:bg-blue-500',
  brand: 'bg-primary-600 dark:bg-primary-400',
}

const CATEGORY_ORDER: TStepStatusCategory[] = [
  'success',
  'error',
  'active',
  'awaiting',
  'pending',
]

export const installStatusKey = (install: TRolloutInstall) =>
  install.awaitingApproval ? 'approval-awaiting' : install.status

export const installStatusClass = (status: string) =>
  stepStatusCategory(status) === 'pending'
    ? 'bg-black/10 dark:bg-white/15'
    : STATUS_FILL[getStatusTheme(status)]

export const InstallStatusTiles = ({
  group,
}: {
  group: TRolloutInstallGroup
}) => {
  const tiles = group.installs
    .map((install) => ({
      id: install.id,
      name: install.name,
      status: installStatusKey(install),
    }))
    .sort(
      (a, b) =>
        CATEGORY_ORDER.indexOf(stepStatusCategory(a.status)) -
        CATEGORY_ORDER.indexOf(stepStatusCategory(b.status))
    )

  if (!tiles.length) return null

  return (
    <span className="flex w-full flex-wrap gap-1">
      {tiles.map((tile) => (
        <Tooltip
          key={tile.id}
          position="top"
          tipContent={
            <span className="flex items-center gap-2">
              <Text variant="subtext">{tile.name}</Text>
              <Status status={tile.status} />
            </span>
          }
        >
          <span
            aria-label={`${tile.name}: ${tile.status}`}
            className={cn(
              'block size-5 rounded-sm',
              installStatusClass(tile.status)
            )}
          />
        </Tooltip>
      ))}
    </span>
  )
}
