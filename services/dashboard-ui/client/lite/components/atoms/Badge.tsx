import type { CSSProperties, HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { Icon } from './Icon'
import { Text } from './Text'

export type TBadgeTone = 'neutral' | 'accent'
export type TBadgeVariant = 'default' | 'code'

export interface IBadge extends HTMLAttributes<HTMLSpanElement> {
  tone?: TBadgeTone
  variant?: TBadgeVariant
  color?: string
  labelKey?: string
  labelValue?: string
  onRemove?: () => void
  removeLabel?: string
  loading?: boolean
  loadingWidth?: number
  disabled?: boolean
}

const TONE_CLASSES: Record<TBadgeTone, string> = {
  neutral: 'bg-badge-bg text-badge-text border-badge-border',
  accent: 'bg-surface-accent text-accent border-divider-accent/40',
}

const SHAPE_CLASSES: Record<TBadgeVariant, string> = {
  default: 'font-sans rounded-full',
  code: 'font-mono rounded-md',
}

const CELL_CLASSES =
  'inline-flex items-center gap-1 border px-2 py-0.5 text-caption'

export const Badge = ({
  tone = 'neutral',
  variant = 'default',
  color,
  labelKey,
  labelValue,
  onRemove,
  removeLabel = 'Remove',
  loading = false,
  loadingWidth,
  disabled = false,
  className,
  children,
  ...props
}: IBadge) => {
  if (loading) {
    return (
      <span
        aria-hidden
        className={cn(
          'skeleton inline-block w-fit px-2 py-0.5 text-caption',
          SHAPE_CLASSES[variant],
          className
        )}
        style={{ width: `${loadingWidth ?? 8}ch` }}
        {...props}
      >
        {'​'}
      </span>
    )
  }

  const remove = onRemove ? (
    <button
      type="button"
      onClick={onRemove}
      disabled={disabled}
      aria-label={`${removeLabel} ${labelKey ?? ''}`.trim()}
      className="-mr-0.5 inline-flex cursor-pointer items-center rounded-xs opacity-70 transition-opacity not-disabled:hover:opacity-100 disabled:cursor-not-allowed disabled:opacity-40 focus-visible:outline-2 focus-visible:outline-offset-1 focus-visible:outline-focus-ring"
    >
      <Icon variant="XIcon" size={12} />
    </button>
  ) : null

  const isLabel = labelKey !== undefined

  if (!isLabel) {
    return (
      <span
        className={cn(
          CELL_CLASSES,
          'w-fit shrink-0',
          SHAPE_CLASSES[variant],
          TONE_CLASSES[tone],
          className
        )}
        {...props}
      >
        {children}
        {remove}
      </span>
    )
  }

  const valueStyle = color
    ? ({ '--badge-color': color } as CSSProperties)
    : undefined

  return (
    <span className={cn('inline-flex w-fit max-w-full', className)} {...props}>
      <span
        className={cn(
          CELL_CLASSES,
          SHAPE_CLASSES[variant],
          TONE_CLASSES.neutral,
          'min-w-0 shrink rounded-r-none'
        )}
      >
        <Text
          variant="caption"
          family={variant === 'code' ? 'mono' : 'sans'}
          className="block min-w-0 truncate"
        >
          {labelKey}
        </Text>
      </span>
      <span
        style={valueStyle}
        className={cn(
          CELL_CLASSES,
          SHAPE_CLASSES[variant],
          color ? 'badge-custom' : TONE_CLASSES[tone],
          'min-w-0 shrink rounded-l-none border-l-0'
        )}
      >
        <Text
          variant="caption"
          family={variant === 'code' ? 'mono' : 'sans'}
          className="block min-w-0 truncate"
        >
          {labelValue}
        </Text>
        {remove}
      </span>
    </span>
  )
}
