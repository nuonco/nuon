import type { HTMLAttributes } from 'react'
import { Icon } from '@/components/common/Icon'
import type { TTheme } from '@/types'
import { cn } from '@/utils/classnames'
import { humanize } from '@/utils/string-utils'
import { getStatusTheme, getStatusIconVariant } from '@/utils/status-utils'
import { SKELETON_CLASSES } from './Skeleton'
import { Text } from './Text'

export type TStatusTheme = TTheme
export type TStatusType = string | 'success' | 'error'
type TStatusVariant = 'default' | 'badge' | 'timeline'

const VARIANT_CLASSES: Record<TStatusVariant, string> = {
  default:
    'text-[12px] font-sans font-strong leading-[17px] tracking-[-0.2px] flex items-center gap-1.5 shrink-0 grow-0 text-cool-grey-950 dark:text-white',
  badge:
    'text-[12px] font-sans font-strong leading-[17px] tracking-[-0.2px] flex items-center gap-1.5 shrink-0 grow-0 text-cool-grey-950 dark:text-white border rounded-full px-2 py-0.5 w-fit',
  timeline:
    'text-[12px] font-sans font-strong leading-[17px] tracking-[-0.2px] flex items-center gap-1.5 shrink-0 grow-0 text-cool-grey-950 dark:text-white',
}

const INDICATOR_BASE =
  'flex items-center justify-center rounded-full shrink-0 grow-0'
const INDICATOR_SIZE: Record<TStatusVariant, string> = {
  default: 'w-1.5 h-1.5',
  badge: 'w-1.5 h-1.5',
  timeline: '',
}

const INDICATOR_THEME_CLASSES: Record<
  TStatusVariant,
  Record<TTheme, string>
> = {
  default: {
    default: 'bg-cool-grey-600 dark:bg-white/70',
    neutral: 'bg-cool-grey-600 dark:bg-white/70',
    success: 'bg-green-600 dark:bg-green-500',
    error: 'bg-red-600 dark:bg-red-500',
    warn: 'bg-orange-600 dark:bg-orange-500',
    info: 'bg-blue-600 dark:bg-blue-500',
    brand: 'bg-primary-600 dark:bg-primary-400',
  },
  badge: {
    default: 'bg-cool-grey-600 dark:bg-white/70',
    neutral: 'bg-cool-grey-600 dark:bg-white/70',
    success: 'bg-green-600 dark:bg-green-500',
    error: 'bg-red-600 dark:bg-red-500',
    warn: 'bg-orange-600 dark:bg-orange-500',
    info: 'bg-blue-600 dark:bg-blue-500',
    brand: 'bg-primary-600 dark:bg-primary-400',
  },
  timeline: {
    default: 'bg-cool-grey-200 dark:bg-cool-grey-800 dark:text-cool-grey-400',
    neutral: 'bg-cool-grey-200 dark:bg-cool-grey-800 dark:text-cool-grey-400',
    success:
      'bg-green-100 text-green-800 dark:bg-green-950 dark:text-green-400',
    error: 'bg-red-100 text-red-800 dark:bg-red-950 dark:text-red-400',
    warn: 'bg-orange-100 text-orange-800 dark:bg-orange-950 dark:text-orange-400',
    info: 'bg-blue-100 text-blue-800 dark:bg-blue-950 dark:text-blue-400',
    brand:
      'bg-primary-200 text-primary-800 dark:bg-primary-950 dark:text-primary-400',
  },
}

const STATUS_TEXT_CLASSES = {
  timeline: 'hidden',
  default: '',
  badge: '',
}

export interface IStatus
  extends Omit<HTMLAttributes<HTMLSpanElement>, 'children'> {
  children?: React.ReactNode
  iconSize?: number
  isWithoutText?: boolean
  loading?: boolean
  loadingWidth?: number
  status?: TStatusType
  statusText?: string
  variant?: TStatusVariant
}

export const Status = ({
  children,
  className,
  iconSize = 18,
  isWithoutText = false,
  loading,
  loadingWidth,
  status,
  variant = 'default',
  ...props
}: IStatus) => {
  if (loading) {
    if (variant === 'timeline') {
      return (
        <span
          aria-hidden
          className={cn(INDICATOR_BASE, SKELETON_CLASSES, className)}
          style={{ width: iconSize + 8, height: iconSize + 8 }}
          {...props}
        />
      )
    }
    return (
      <span
        aria-hidden
        className={cn(VARIANT_CLASSES[variant], className)}
        {...props}
      >
        <span
          className={cn(INDICATOR_BASE, INDICATOR_SIZE[variant], SKELETON_CLASSES)}
        />
        {isWithoutText ? null : (
          <Text variant="subtext" loading loadingWidth={loadingWidth} />
        )}
      </span>
    )
  }

  const theme = getStatusTheme(status ?? '')
  const iconVariant =
    variant === 'timeline' ? getStatusIconVariant(status ?? '') : null

  const rootClass = cn(VARIANT_CLASSES[variant], className)
  const indicatorClass = cn(
    INDICATOR_BASE,
    INDICATOR_SIZE[variant],
    INDICATOR_THEME_CLASSES[variant][theme],
  )
  const indicatorStyle =
    variant === 'timeline'
      ? { width: iconSize + 8, height: iconSize + 8 }
      : undefined
  const statusTextClass = STATUS_TEXT_CLASSES[variant]

  return (
    <span className={rootClass} {...props}>
      <span className={indicatorClass} style={indicatorStyle}>
        {iconVariant ? (
          <Icon
            className="status-icon"
            variant={iconVariant}
            weight="bold"
            size={iconSize}
          />
        ) : null}
      </span>

      {isWithoutText ? null : (
        <span className={statusTextClass}>
          {typeof children === 'string'
            ? humanize(children)
            : children || humanize(status || 'unknown')}
        </span>
      )}
    </span>
  )
}
