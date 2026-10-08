import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import {
  installStatusClass,
  sortedStatusInstalls,
  type TStatusInstall,
} from './install-status'

export const InstallStatusTiles = ({
  installs,
}: {
  installs: TStatusInstall[]
}) => {
  const tiles = sortedStatusInstalls(installs)
  if (!tiles.length) return null
  if (tiles.length > 80) return null

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
