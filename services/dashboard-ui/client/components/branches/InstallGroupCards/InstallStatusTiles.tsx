import { Link } from '@/components/common/Link'
import { Status } from '@/components/common/Status'
import { Text } from '@/components/common/Text'
import { Tooltip } from '@/components/common/Tooltip'
import { cn } from '@/utils/classnames'
import {
  installStatusBucket,
  installStatusClass,
  sortedStatusInstalls,
  type TStatusInstall,
} from './install-status'

const segmentLabel = (tile: TStatusInstall) => `${tile.name}: ${tile.status}`

const hitClass = (variant: 'tiles' | 'rail') =>
  cn(
    'cursor-pointer rounded-sm hover:opacity-80 focus-visible:outline-none focus-visible:shadow-[0_0_0_1px_var(--background),0_0_0_3px_var(--focus-ring)]',
    variant === 'tiles' ? 'block size-5' : 'flex h-6 w-full items-center'
  )

const barClass = (tile: TStatusInstall, variant: 'tiles' | 'rail') =>
  cn(
    'block w-full rounded-sm',
    variant === 'tiles'
      ? 'h-full'
      : installStatusBucket(tile.status) === 'in-progress'
        ? 'h-2'
        : 'h-1',
    installStatusClass(tile.status)
  )

export const InstallStatusTiles = ({
  installs,
  variant = 'tiles',
  onSelectInstall,
}: {
  installs: TStatusInstall[]
  variant?: 'tiles' | 'rail'
  onSelectInstall?: (install: TStatusInstall) => void
}) => {
  const tiles = sortedStatusInstalls(installs)
  if (!tiles.length) return null
  if (tiles.length > 80) return null

  return (
    <span
      className={cn(
        'flex w-full gap-1',
        variant === 'tiles' ? 'flex-wrap' : 'min-w-0'
      )}
      aria-label={variant === 'rail' ? 'Install progress' : undefined}
    >
      {tiles.map((tile) => {
        const label = segmentLabel(tile)
        const opensDrawer = Boolean(onSelectInstall)
        return (
          <Tooltip
            key={tile.id}
            position="top"
            className={variant === 'rail' ? 'block min-w-0 flex-1' : undefined}
            style={variant === 'rail' ? { width: '100%' } : undefined}
            tipContent={
              <span className="flex items-center gap-2">
                <Text variant="subtext">{tile.name}</Text>
                <Status status={tile.status} />
              </span>
            }
          >
            {opensDrawer && onSelectInstall ? (
              <button
                type="button"
                aria-label={label}
                className={hitClass(variant)}
                onClick={() => onSelectInstall(tile)}
              >
                <span className={barClass(tile, variant)} />
              </button>
            ) : variant === 'rail' && tile.overviewHref ? (
              // Square status tiles stay inert. Only the progress rail links through.
              <Link
                href={tile.overviewHref}
                aria-label={label}
                className={cn(
                  hitClass(variant),
                  '!w-full !font-inherit !leading-none !tracking-normal !text-inherit !no-underline !focus-visible:outline-none'
                )}
              >
                <span className={barClass(tile, variant)} />
              </Link>
            ) : (
              <span
                aria-label={label}
                className={cn(
                  'block rounded-sm',
                  variant === 'tiles'
                    ? 'size-5'
                    : installStatusBucket(tile.status) === 'in-progress'
                      ? 'h-2 w-full'
                      : 'h-1 w-full',
                  installStatusClass(tile.status)
                )}
              />
            )}
          </Tooltip>
        )
      })}
    </span>
  )
}
