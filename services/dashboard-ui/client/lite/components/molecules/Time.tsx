import { DateTime, type DateTimeFormatOptions } from 'luxon'
import type { ReactNode } from 'react'
import { useNow } from '../../hooks/use-now'
import type { TPopoverSide } from '../../hooks/use-popover'
import { Text, type IText } from '../atoms/Text'
import { Tooltip } from '../atoms/Tooltip'

export type TTimeFormat =
  | 'relative'
  | 'datetime'
  | 'full'
  | 'date'
  | 'time'
  | 'precise'

export interface ITime extends Omit<IText, 'as' | 'children'> {
  value?: string | number
  format?: TTimeFormat
  tooltip?: boolean
  tooltipSide?: TPopoverSide
  live?: boolean
  fallback?: ReactNode
}

const FULL_FORMAT: DateTimeFormatOptions = {
  ...DateTime.DATETIME_FULL_WITH_SECONDS,
  timeZoneName: 'short',
}

const formatRelative = (value: DateTime, now: number) => {
  const seconds = Math.abs(
    DateTime.fromMillis(now).diff(value, 'seconds').seconds
  )
  return seconds < 10
    ? 'just now'
    : value.toRelative({ base: DateTime.fromMillis(now) })
}

const formatTime = (value: DateTime, format: TTimeFormat, now: number) => {
  switch (format) {
    case 'relative':
      return formatRelative(value, now)
    case 'full':
      return value.toLocaleString(FULL_FORMAT)
    case 'date':
      return value.toLocaleString(DateTime.DATE_MED)
    case 'time':
      return value.toLocaleString(DateTime.TIME_SIMPLE)
    case 'precise':
      return value.toFormat('yyyy-LL-dd HH:mm:ss.SSS ZZZZ')
    case 'datetime':
    default:
      return value.toLocaleString(DateTime.DATETIME_MED)
  }
}

const ABSOLUTE_FORMATS: TTimeFormat[] = ['full', 'precise']

export const Time = ({
  value,
  format = 'datetime',
  tooltip = true,
  tooltipSide = 'top',
  live = format === 'relative',
  fallback = '—',
  loading = false,
  loadingWidth,
  variant = 'caption',
  color = 'secondary',
  family,
  className,
  ...props
}: ITime) => {
  const tooltipIsRelative = ABSOLUTE_FORMATS.includes(format)
  const now = useNow(
    (live && format === 'relative') || (tooltip && tooltipIsRelative)
  )
  const parsed =
    typeof value === 'number'
      ? DateTime.fromSeconds(value)
      : typeof value === 'string'
        ? DateTime.fromISO(value)
        : undefined
  const valid = parsed?.isValid ? parsed : undefined

  if (loading) {
    return (
      <Text
        variant={variant}
        color={color}
        family={family}
        loading
        loadingWidth={loadingWidth}
        className={className}
        {...props}
      />
    )
  }

  if (!valid) {
    return (
      <Text
        variant={variant}
        color="tertiary"
        family={family}
        className={className}
        {...props}
      >
        {fallback}
      </Text>
    )
  }

  const content = (
    <time dateTime={valid.toISO() ?? undefined} className="inline-flex">
      <Text
        variant={variant}
        color={color}
        family={family ?? (format === 'precise' ? 'mono' : 'sans')}
        className={className}
        {...props}
      >
        {formatTime(valid, format, now)}
      </Text>
    </time>
  )

  if (!tooltip) return content

  return (
    <Tooltip
      side={tooltipSide}
      content={
        tooltipIsRelative
          ? formatRelative(valid, now)
          : valid.toLocaleString(FULL_FORMAT)
      }
    >
      {content}
    </Tooltip>
  )
}
