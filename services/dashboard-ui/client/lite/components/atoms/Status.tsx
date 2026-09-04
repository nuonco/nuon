import type { CSSProperties, HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import { getStatusTheme, type TStatusTheme } from '@/utils/status-utils'
import { Icon, type TIconVariant } from './Icon'
import { Spinner } from './Spinner'
import { Text } from './Text'

export type TStatusVariant = 'chip' | 'inline' | 'dot'

export interface IStatus extends HTMLAttributes<HTMLSpanElement> {
  status: string
  label?: string
  variant?: TStatusVariant
  theme?: TStatusTheme
  loading?: boolean
  loadingWidth?: number
}

const THEME_VAR: Record<TStatusTheme, string> = {
  success: 'var(--status-success)',
  error: 'var(--status-error)',
  warn: 'var(--status-warn)',
  info: 'var(--status-info)',
  brand: 'var(--status-brand)',
  neutral: 'var(--status-neutral)',
}

const THEME_ICON: Record<TStatusTheme, TIconVariant> = {
  success: 'CheckCircleIcon',
  error: 'XCircleIcon',
  warn: 'WarningIcon',
  info: 'ClockCountdownIcon',
  brand: 'SparkleIcon',
  neutral: 'ClockCountdownIcon',
}

const humanize = (status: string) =>
  status.replace(/[-_\s]+/g, ' ').replace(/^./, (c) => c.toUpperCase())

export const Status = ({
  status,
  label,
  variant = 'chip',
  theme,
  loading = false,
  loadingWidth,
  className,
  ...props
}: IStatus) => {
  const resolved = theme ?? getStatusTheme(status)
  const text = label ?? humanize(status)
  const style = { '--status-color': THEME_VAR[resolved] } as CSSProperties

  if (loading) {
    return (
      <Text
        variant="caption"
        loading
        loadingWidth={loadingWidth ?? 10}
        className={className}
      />
    )
  }

  const glyph =
    resolved === 'info' ? (
      <Spinner size={19} />
    ) : (
      <Icon variant={THEME_ICON[resolved]} size={19} />
    )

  if (variant === 'dot') {
    return (
      <span
        style={style}
        className={cn('inline-flex w-fit items-center', className)}
        {...props}
      >
        <span
          aria-hidden
          className="size-2 rounded-full"
          style={{ backgroundColor: 'var(--status-color)' }}
        />
        <span className="sr-only">{text}</span>
      </span>
    )
  }

  if (variant === 'inline') {
    return (
      <span
        style={style}
        className={cn('inline-flex w-fit items-center gap-1.5', className)}
        {...props}
      >
        <span aria-hidden style={{ color: 'var(--status-color)' }} className="flex">
          {glyph}
        </span>
        <Text variant="caption">{text}</Text>
      </span>
    )
  }

  return (
    <span
      style={style}
      className={cn(
        'status-tint inline-flex w-fit shrink-0 items-center gap-1.5 rounded-md px-1.5 py-0.5 font-medium',
        className
      )}
      {...props}
    >
      <span aria-hidden className="flex">
        {glyph}
      </span>
      <Text variant="caption" className="block">
        {text}
      </Text>
    </span>
  )
}
