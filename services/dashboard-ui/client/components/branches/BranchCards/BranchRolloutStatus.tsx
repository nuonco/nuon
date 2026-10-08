import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import { getStatusTheme, type TStatusTheme } from '@/utils/status-utils'
import { humanize } from '@/utils/string-utils'

export type TBranchRolloutInstall = {
  id: string
  name: string
  status: string
}

const MAX_SQUARES = 12

const SQUARE_THEME_CLASSES: Record<TStatusTheme, string> = {
  success: 'bg-green-600 dark:bg-green-500',
  error: 'bg-red-600 dark:bg-red-500',
  warn: 'bg-orange-600 dark:bg-orange-500',
  info: 'bg-blue-600 dark:bg-blue-500',
  brand: 'bg-primary-600 dark:bg-primary-400',
  neutral: 'bg-cool-grey-300 dark:bg-dark-grey-600',
}

const InstallCount = ({
  count,
  isPartial,
}: {
  count: number
  isPartial?: boolean
}) => (
  <Text variant="subtext" theme="neutral">
    {count}
    {isPartial ? '+' : ''} {count === 1 && !isPartial ? 'install' : 'installs'}
  </Text>
)

const statusCounts = (installs: TBranchRolloutInstall[]) => {
  const counts = new Map<string, number>()
  for (const install of installs) {
    counts.set(install.status, (counts.get(install.status) ?? 0) + 1)
  }
  return [...counts.entries()].sort((a, b) => b[1] - a[1])
}

export const BranchRolloutStatus = ({
  installs = [],
  installCount,
  installCountIsPartial,
}: {
  installs?: TBranchRolloutInstall[]
  installCount?: number
  installCountIsPartial?: boolean
}) => {
  if (installs.length === 0 && installCount === undefined) return null

  const visible =
    installs.length > MAX_SQUARES ? installs.slice(0, MAX_SQUARES) : installs
  const hidden = installs.length - visible.length

  return (
    <div className="flex flex-col gap-2 min-w-0">
      {installs.length > 0 ? (
        <div className="flex flex-wrap gap-1">
          {visible.map((install) => (
            <Tooltip
              key={install.id}
              position="top"
              tipContent={
                <span className="flex items-center gap-2">
                  <Text variant="subtext">{install.name}</Text>
                  <Status status={install.status} />
                </span>
              }
            >
              <span
                aria-label={`${install.name}: ${humanize(install.status)}`}
                className={cn(
                  'block size-6 rounded-sm',
                  SQUARE_THEME_CLASSES[getStatusTheme(install.status)]
                )}
              />
            </Tooltip>
          ))}
          {hidden > 0 ? (
            <span className="flex h-6 items-center justify-center rounded-sm bg-cool-grey-200 px-1.5 text-xs text-cool-grey-700 dark:bg-dark-grey-700 dark:text-cool-grey-300">
              {hidden} more
            </span>
          ) : null}
        </div>
      ) : null}
      <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
        {installCount !== undefined ? (
          <InstallCount count={installCount} isPartial={installCountIsPartial} />
        ) : null}
        {statusCounts(installs).map(([status, count]) => (
          <Status key={status} status={status}>
            <span className="font-normal">
              {count} {humanize(status).toLowerCase()}
            </span>
          </Status>
        ))}
      </div>
    </div>
  )
}
