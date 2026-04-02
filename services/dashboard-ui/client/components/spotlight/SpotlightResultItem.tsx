import { Icon } from '@/components/common/Icon'
import { Badge } from '@/components/common/Badge'
import { cn } from '@/utils/classnames'
import type { SpotlightResult } from './types'

interface ISpotlightResultItem {
  result: SpotlightResult
  isActive: boolean
  index: number
  onSelect: () => void
  onHover: () => void
}

export const SpotlightResultItem = ({
  result,
  isActive,
  index,
  onSelect,
  onHover,
}: ISpotlightResultItem) => (
  <button
    key={result.path ?? result.label}
    data-index={index}
    className={cn(
      'transition duration-200 px-2 py-1 cursor-pointer select-none rounded text-sm text-left flex items-center gap-3',
      {
        'text-white bg-primary-600': isActive,
        'hover:bg-black/5 dark:hover:bg-white/5': !isActive,
      }
    )}
    onClick={onSelect}
    onMouseEnter={onHover}
  >
    <Icon
      variant={result.icon}
      className={cn('shrink-0', {
        'text-white': isActive,
        'text-cool-grey-700 dark:text-cool-grey-500': !isActive,
      })}
    />
    <div className="flex flex-col min-w-0 flex-1">
      <span className="truncate">{result.label}</span>
      {result.subtitle && (
        <span
          className={cn('text-xs truncate', {
            'text-white/70': isActive,
            'text-cool-grey-500': !isActive,
          })}
        >
          {result.subtitle}
        </span>
      )}
    </div>
    {result.tag && (
      <Badge size="sm" variant="code" theme={result.tag === 'command' ? 'brand' : 'neutral'} className="shrink-0">
        {result.tag}
      </Badge>
    )}
  </button>
)
