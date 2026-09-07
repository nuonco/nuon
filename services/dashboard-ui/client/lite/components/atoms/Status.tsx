import type { CSSProperties, HTMLAttributes } from 'react'
import { cn } from '@/utils/classnames'
import type { TCompositeStatus } from '@/types/ctl-api.types'
import {
  getStatusIconVariant,
  getStatusTheme,
  type TStatusTheme,
} from '@/utils/status-utils'
import { Icon, type TIconVariant } from './Icon'
import { Spinner } from './Spinner'
import { Text } from './Text'
import { Tooltip } from './Tooltip'

export type TStatusVariant = 'chip' | 'inline' | 'dot' | 'icon'

export interface IStatus extends HTMLAttributes<HTMLSpanElement> {
  status?: string | TCompositeStatus
  label?: string
  description?: string
  icon?: TIconVariant
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
  description,
  icon,
  variant = 'chip',
  theme,
  loading = false,
  loadingWidth,
  className,
  ...props
}: IStatus) => {
  const statusValue =
    typeof status === 'string' ? status : (status?.status ?? 'unknown')
  const statusDescription =
    description ??
    (typeof status === 'object' ? status?.status_human_description : undefined)
  const resolved = theme ?? getStatusTheme(statusValue)
  const resolvedIcon = icon ?? getStatusIconVariant(statusValue)
  const text = label ?? humanize(statusValue)
  const style = { '--status-color': THEME_VAR[resolved] } as CSSProperties

  if (loading) {
    if (variant === 'icon') {
      return (
        <span
          aria-hidden
          className={cn('skeleton size-6 rounded-full', className)}
        />
      )
    }

    return (
      <Text
        variant="caption"
        loading
        loadingWidth={loadingWidth ?? 10}
        className={className}
      />
    )
  }

  const glyphAt = (size: number) =>
    resolvedIcon === 'Loading' ? (
      <Spinner size={size} />
    ) : resolvedIcon === 'none' ? null : (
      <Icon variant={resolvedIcon} size={size} />
    )

  const tooltipGlyph =
    resolvedIcon === 'Loading' || resolvedIcon === 'none' ? (
      <Icon variant={THEME_ICON[resolved]} size={16} />
    ) : (
      <Icon variant={resolvedIcon} size={16} />
    )

  const glyph = glyphAt(19)

  const descriptionText = statusDescription ? (
    <Text
      as="p"
      variant="caption"
      color="secondary"
      lines={10}
      className="break-words [overflow-wrap:anywhere]"
    >
      {statusDescription}
    </Text>
  ) : null

  const titledTooltip = (
    <span className="flex max-w-sm flex-col gap-1 whitespace-normal text-left">
      <span className="flex items-center gap-1.5">
        <span
          aria-hidden
          className="flex"
          style={{ color: THEME_VAR[resolved] }}
        >
          {tooltipGlyph}
        </span>
        <Text variant="caption" weight="semibold">
          {text}
        </Text>
      </span>
      {descriptionText}
    </span>
  )

  if (variant === 'dot' || variant === 'icon') {
    return (
      <Tooltip
        content={titledTooltip}
        contentClassName="max-w-sm"
        tabIndex={0}
        aria-label={text}
        style={style}
        className={cn('items-center', className)}
        {...props}
      >
        {variant === 'icon' ? (
          <span
            aria-hidden
            className="status-tint flex size-6 shrink-0 items-center justify-center rounded-full"
          >
            {glyphAt(16)}
          </span>
        ) : (
          <span
            aria-hidden
            className={cn(
              'size-2 rounded-full',
              resolved === 'info' && 'status-pulse'
            )}
            style={{ backgroundColor: 'var(--status-color)' }}
          />
        )}
      </Tooltip>
    )
  }

  if (variant === 'inline') {
    const content = (
      <span
        style={style}
        className={cn('inline-flex w-fit items-center gap-1.5', className)}
        {...props}
      >
        <span
          aria-hidden
          style={{ color: 'var(--status-color)' }}
          className="flex"
        >
          {glyph}
        </span>
        <Text variant="caption">{text}</Text>
      </span>
    )

    if (!statusDescription) return content

    return (
      <Tooltip content={titledTooltip} contentClassName="max-w-sm" tabIndex={0}>
        {content}
      </Tooltip>
    )
  }

  const content = (
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

  if (!statusDescription) return content

  return (
    <Tooltip content={titledTooltip} contentClassName="max-w-sm" tabIndex={0}>
      {content}
    </Tooltip>
  )
}
